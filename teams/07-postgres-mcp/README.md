# 07 - Postgres MCP

A custom MCP server backed by a real Postgres database. Three agents collaborate
over live data: a reader explores the schema, an analyst surfaces patterns, and
an auditor hunts security issues - all by writing SQL queries via MCP tools.

No mock data. No third-party APIs. Everything runs in-cluster.

## What's included

```
secret.yaml        Ollama config + Postgres credentials
postgres.yaml      Postgres Deployment + Service + seed SQL (ConfigMap)
postgres-mcp.yaml  MCP server Deployment + Service
data-team.yaml     ArkTeam: reader → analyst + auditor (parallel)
src/
  main.go          Go MCP server; connects to Postgres, serves 3 tools
  go.mod           One dependency: github.com/lib/pq
  Dockerfile       Two-stage build; distroless runtime
```

## Database schema

Three tables seeded with realistic data including deliberate security issues:

**users** - 8 rows
- Dormant admin account (last login 8 months ago, still active)
- User with 14 consecutive failed login attempts
- Ghost account (never logged in) with a pending order

**orders** - 8 rows
- $15,000 order from the ghost account
- Large historical order from the dormant admin

**api_keys** - 4 rows
- Superadmin key with no expiry date
- Admin key attached to the dormant account
- Expired key (30 days ago) with a `last_used` timestamp of 2 hours ago

## MCP tools

| Tool | Description |
|------|-------------|
| `list_tables` | List all tables in the public schema |
| `describe_table` | Column names, types, nullability |
| `run_query` | Execute any SELECT; max 100 rows returned |

`run_query` rejects non-SELECT statements at the server level.

## Pipeline

```
reader ──┬── analyst   (data patterns, business metrics)
         └── auditor   (security findings with severity)
```

`analyst` and `auditor` run in parallel after `reader` completes.

## Apply

```bash
# 1. Build the MCP server image
docker build -t postgres-mcp:latest ./src

# 2. Load into kind
kind load docker-image postgres-mcp:latest --name ark

# 3. Deploy (order matters: postgres before postgres-mcp)
kubectl apply -f secret.yaml
kubectl apply -f postgres.yaml
kubectl wait --for=condition=ready pod -l app=postgres -n ark-teams --timeout=60s
kubectl apply -f postgres-mcp.yaml
kubectl wait --for=condition=ready pod -l app=postgres-mcp -n ark-teams --timeout=60s
kubectl apply -f data-team.yaml

# 4. Wait for agent pods
kubectl wait --for=condition=ready pod \
  -l app.kubernetes.io/managed-by=ark -n ark-teams --timeout=120s
```

## Trigger

```bash
ark trigger data-team -n ark-teams
```

## Expected output

```
DATA INSIGHTS
=============
Users: 8 total — 2 admins, 1 superadmin, 5 regular users. 1 disabled account.
Orders: 8 total — $32,149.98 total value. 4 pending, 3 completed, 1 failed.
API Keys: 4 total — 2 with no expiry date. 1 key expired.

SECURITY REPORT
===============
[CRITICAL] Expired API key still in use
  Key ID 4 (frank@corp.com) expired 30 days ago but was used 2 hours ago.
  Action: revoke key immediately and audit recent API activity.

[CRITICAL] Ghost account with high-value order
  eve@corp.com has never logged in but placed a $15,000 pending order.
  Action: flag order for manual review; require login verification.

[HIGH] Dormant privileged account
  alice@corp.com (admin) has not logged in for 8 months and remains active.
  Action: disable account or require re-authentication.
...
```

## Adapting to your own database

Point the MCP server at any Postgres-compatible DB by changing the env vars
in `postgres-mcp.yaml`. The tools, agent prompts, and team pipeline stay the same.
