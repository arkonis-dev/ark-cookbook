// incident-mcp — a minimal MCP HTTP server exposing three tools:
//
//   list_incidents       returns all active incidents (read)
//   get_incident_details returns full details for one incident (read)
//   escalate_incident    marks an incident as escalated (write)
//
// The ark MCP client uses a simple HTTP transport (not SSE):
//   POST /tools/list  → {"tools": [...]}
//   POST /tools/call  → {"name":"...", "arguments":{...}} → {"content":[{"type":"text","text":"..."}]}
//   GET  /health      → 200
//
// The server has no external dependencies — pure Go stdlib, mock data in-memory.
// Swap the mock store for a real ticketing API (PagerDuty, Jira, OpsGenie) and re-deploy.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// Incident store (mock data — replace with real API calls)
// ---------------------------------------------------------------------------

type incident struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Severity    string    `json:"severity"`
	Status      string    `json:"status"`
	Service     string    `json:"service"`
	OpenedAt    time.Time `json:"opened_at"`
	Description string    `json:"description"`
	Escalated   bool      `json:"escalated"`
}

var (
	incMu     sync.RWMutex
	incidents = []incident{
		{
			ID: "INC-001", Title: "API gateway 5xx spike", Severity: "critical",
			Status: "investigating", Service: "api-gateway",
			OpenedAt: time.Now().Add(-47 * time.Minute),
			Description: "Error rate on /api/v2 climbed from 0.1% to 14% at 14:32 UTC. " +
				"All regions affected. No recent deploys. Upstream dependency latency nominal.",
		},
		{
			ID: "INC-002", Title: "Postgres replica lag > 30s", Severity: "high",
			Status: "investigating", Service: "user-db",
			OpenedAt: time.Now().Add(-22 * time.Minute),
			Description: "Read replica falling 34 seconds behind primary. " +
				"Root cause suspected: long-running analytics query holding a lock. " +
				"Write traffic unaffected; read endpoints degraded.",
		},
		{
			ID: "INC-003", Title: "Payment webhook queue backlog", Severity: "medium",
			Status: "acknowledged", Service: "payments",
			OpenedAt: time.Now().Add(-3 * time.Hour),
			Description: "Stripe webhook processing queue at 4,800 unprocessed events. " +
				"Consumer throughput dropped after v2.4.1 deploy. Payments complete; " +
				"downstream reconciliation delayed by ~6 hours.",
		},
	}
)

func doListIncidents() interface{} {
	incMu.RLock()
	defer incMu.RUnlock()
	type row struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Severity  string `json:"severity"`
		Status    string `json:"status"`
		Service   string `json:"service"`
		AgeMin    int    `json:"age_minutes"`
		Escalated bool   `json:"escalated"`
	}
	rows := make([]row, len(incidents))
	for i, inc := range incidents {
		rows[i] = row{
			ID: inc.ID, Title: inc.Title, Severity: inc.Severity,
			Status: inc.Status, Service: inc.Service,
			AgeMin: int(time.Since(inc.OpenedAt).Minutes()), Escalated: inc.Escalated,
		}
	}
	return map[string]interface{}{"incidents": rows, "count": len(rows)}
}

func doGetDetails(id string) (interface{}, error) {
	incMu.RLock()
	defer incMu.RUnlock()
	for _, inc := range incidents {
		if inc.ID == id {
			return map[string]interface{}{
				"id": inc.ID, "title": inc.Title, "severity": inc.Severity,
				"status": inc.Status, "service": inc.Service,
				"age_minutes": int(time.Since(inc.OpenedAt).Minutes()),
				"description": inc.Description, "escalated": inc.Escalated,
			}, nil
		}
	}
	return nil, fmt.Errorf("incident %s not found", id)
}

func doEscalate(id string) (interface{}, error) {
	incMu.Lock()
	defer incMu.Unlock()
	for i, inc := range incidents {
		if inc.ID == id {
			if inc.Escalated {
				return map[string]string{"status": "already_escalated", "id": id}, nil
			}
			incidents[i].Escalated = true
			incidents[i].Status = "escalated"
			log.Printf("incident %s escalated", id)
			return map[string]string{"status": "escalated", "id": id}, nil
		}
	}
	return nil, fmt.Errorf("incident %s not found", id)
}

// ---------------------------------------------------------------------------
// Tool definitions
// ---------------------------------------------------------------------------

var toolDefs = []map[string]interface{}{
	{
		"name":        "list_incidents",
		"description": "List all active incidents with severity, status, service, and age in minutes.",
		"inputSchema": map[string]interface{}{
			"type": "object", "properties": map[string]interface{}{},
		},
	},
	{
		"name":        "get_incident_details",
		"description": "Get full details and description for a specific incident by ID.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "string", "description": "Incident ID, e.g. INC-001",
				},
			},
			"required": []string{"id"},
		},
	},
	{
		"name":        "escalate_incident",
		"description": "Mark an incident as escalated. Use when severity is critical and open > 30 minutes.",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "string", "description": "Incident ID to escalate, e.g. INC-001",
				},
			},
			"required": []string{"id"},
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

func toolsCallHandler(w http.ResponseWriter, r *http.Request) {
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
	case "list_incidents":
		result = doListIncidents()
	case "get_incident_details":
		var a struct {
			ID string `json:"id"`
		}
		if jsonErr := json.Unmarshal(req.Arguments, &a); jsonErr != nil || a.ID == "" {
			http.Error(w, `{"error":"missing id"}`, http.StatusBadRequest)
			return
		}
		result, err = doGetDetails(a.ID)
	case "escalate_incident":
		var a struct {
			ID string `json:"id"`
		}
		if jsonErr := json.Unmarshal(req.Arguments, &a); jsonErr != nil || a.ID == "" {
			http.Error(w, `{"error":"missing id"}`, http.StatusBadRequest)
			return
		}
		result, err = doEscalate(a.ID)
	default:
		http.Error(w, fmt.Sprintf(`{"error":"unknown tool %q"}`, req.Name), http.StatusNotFound)
		return
	}

	if err != nil {
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

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/tools/list", toolsListHandler)
	mux.HandleFunc("/tools/call", toolsCallHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Println("incident-mcp listening on :8099")
	srv := &http.Server{
		Addr:              ":8099",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
