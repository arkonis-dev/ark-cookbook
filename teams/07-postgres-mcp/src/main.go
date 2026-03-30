// postgres-mcp — MCP HTTP server that exposes a Postgres database as agent tools.
//
// Tools:
//   list_tables    → list all tables in the public schema
//   describe_table → column names, types, nullability for one table
//   run_query      → execute a SELECT query; returns rows as JSON (max 100 rows)
//
// Transport (ark MCP HTTP protocol):
//   POST /tools/list   → {"tools":[...]}
//   POST /tools/call   → {"name":"...","arguments":{...}} → {"content":[{"type":"text","text":"..."}]}
//   GET  /health       → 200
//
// Config (env vars):
//   DB_HOST      default "postgres"
//   DB_PORT      default "5432"
//   DB_NAME      default "appdb"
//   DB_USER      default "appuser"
//   DB_PASSWORD  required
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// ---------------------------------------------------------------------------
// Database
// ---------------------------------------------------------------------------

func openDB() (*sql.DB, error) {
	host := envOr("DB_HOST", "postgres")
	port := envOr("DB_PORT", "5432")
	name := envOr("DB_NAME", "appdb")
	user := envOr("DB_USER", "appuser")
	pass := os.Getenv("DB_PASSWORD")
	if pass == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		host, port, name, user, pass)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ---------------------------------------------------------------------------
// Tools
// ---------------------------------------------------------------------------

var tableNameRE = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func doListTables(db *sql.DB) (interface{}, error) {
	rows, err := db.Query(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		ORDER BY table_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return map[string]interface{}{"tables": tables}, rows.Err()
}

func doDescribeTable(db *sql.DB, table string) (interface{}, error) {
	if !tableNameRE.MatchString(table) {
		return nil, fmt.Errorf("invalid table name %q", table)
	}
	rows, err := db.Query(`
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1
		ORDER BY ordinal_position`, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type col struct {
		Name     string  `json:"column_name"`
		Type     string  `json:"data_type"`
		Nullable string  `json:"is_nullable"`
		Default  *string `json:"column_default,omitempty"`
	}
	var cols []col
	for rows.Next() {
		var c col
		if err := rows.Scan(&c.Name, &c.Type, &c.Nullable, &c.Default); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("table %q not found", table)
	}
	return map[string]interface{}{"table": table, "columns": cols}, rows.Err()
}

func doRunQuery(db *sql.DB, query string) (interface{}, error) {
	upper := strings.TrimSpace(strings.ToUpper(query))
	if !strings.HasPrefix(upper, "SELECT") {
		return nil, fmt.Errorf("only SELECT queries are allowed")
	}

	// Wrap in a limit to protect context window.
	limited := fmt.Sprintf("SELECT * FROM (%s) _q LIMIT 100", query)

	rows, err := db.Query(limited)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for rows.Next() {
		ptrs := make([]interface{}, len(cols))
		vals := make([]interface{}, len(cols))
		for i := range ptrs {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(map[string]interface{}, len(cols))
		for i, col := range cols {
			row[col] = jsonSafe(vals[i])
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"rows":  result,
		"count": len(result),
	}, nil
}

// jsonSafe converts postgres driver values to JSON-friendly types.
func jsonSafe(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case []byte:
		return string(t)
	case time.Time:
		return t.Format(time.RFC3339)
	default:
		return t
	}
}

// ---------------------------------------------------------------------------
// Tool definitions
// ---------------------------------------------------------------------------

var toolDefs = []map[string]interface{}{
	{
		"name":        "list_tables",
		"description": "List all tables in the database.",
		"inputSchema": map[string]interface{}{
			"type": "object", "properties": map[string]interface{}{},
		},
	},
	{
		"name":        "describe_table",
		"description": "Return column names, data types, and nullability for a table.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"table": map[string]interface{}{
					"type": "string", "description": "Table name, e.g. users",
				},
			},
			"required": []string{"table"},
		},
	},
	{
		"name":        "run_query",
		"description": "Execute a read-only SELECT query and return results as JSON. Maximum 100 rows returned. Use this for counts, aggregations, filtered lookups, and joins.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"sql": map[string]interface{}{
					"type": "string", "description": "SQL SELECT statement to execute",
				},
			},
			"required": []string{"sql"},
		},
	},
}

// ---------------------------------------------------------------------------
// HTTP handlers
// ---------------------------------------------------------------------------

func toolsListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"tools": toolDefs})
}

func toolsCallHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var (
			result interface{}
			err    error
		)

		switch req.Name {
		case "list_tables":
			result, err = doListTables(db)

		case "describe_table":
			var a struct {
				Table string `json:"table"`
			}
			if jsonErr := json.Unmarshal(req.Arguments, &a); jsonErr != nil || a.Table == "" {
				writeErr(w, "missing table argument", http.StatusBadRequest)
				return
			}
			result, err = doDescribeTable(db, a.Table)

		case "run_query":
			var a struct {
				SQL string `json:"sql"`
			}
			if jsonErr := json.Unmarshal(req.Arguments, &a); jsonErr != nil || a.SQL == "" {
				writeErr(w, "missing sql argument", http.StatusBadRequest)
				return
			}
			result, err = doRunQuery(db, a.SQL)

		default:
			writeErr(w, fmt.Sprintf("unknown tool %q", req.Name), http.StatusNotFound)
			return
		}

		if err != nil {
			log.Printf("tool %s error: %v", req.Name, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		text, _ := json.Marshal(result)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": string(text)},
			},
		})
	}
}

func writeErr(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	db, err := openDB()
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	// Wait for postgres to be ready.
	for i := 0; i < 10; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		log.Printf("waiting for postgres... (%d/10)", i+1)
		time.Sleep(2 * time.Second)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("postgres not reachable: %v", err)
	}
	log.Println("connected to postgres")

	mux := http.NewServeMux()
	mux.HandleFunc("/tools/list", toolsListHandler)
	mux.HandleFunc("/tools/call", toolsCallHandler(db))
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		if err := db.Ping(); err != nil {
			http.Error(w, "db unreachable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	log.Println("postgres-mcp listening on :8099")
	srv := &http.Server{
		Addr:              ":8099",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
