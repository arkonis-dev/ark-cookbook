# 06 - Custom MCP Server

Build and deploy your own MCP tool server, then wire it into an ArkTeam pipeline.
The example ships a minimal Go HTTP server that exposes three incident management
tools over the MCP SSE protocol - no external dependencies, no API keys for the
server itself.

## What you build

```
src/
  main.go      Go MCP SSE server — zero external deps, pure stdlib
  Dockerfile   Two-stage build; distroless runtime image
incident-mcp.yaml   Kubernetes Deployment + Service for the MCP server
incident-team.yaml  Three-step ArkTeam pipeline that calls the server
```

The MCP server exposes:

| Tool | Type | Description |
|------|------|-------------|
| `list_incidents` | read | All active incidents with severity, status, age |
| `get_incident_details` | read | Full description for a specific incident |
| `escalate_incident` | write | Mark an incident escalated (mutates server state) |

The pipeline:
1. **triage** - calls `list_incidents` + `get_incident_details`, returns JSON
2. **runbook** - writes a diagnostic runbook for the most critical incident
3. **escalation** - applies escalation criteria, calls `escalate_incident` if warranted

## Features demonstrated

- `spec.roles[*].mcpServers` - connecting agent roles to a custom MCP server URL
- Read tools and write/action tools in the same server
- Chaining tool-calling steps (triage output feeds both runbook and escalation)
- `outputSchema` validation on the triage step
- Deploying a custom MCP server alongside agent pods in the same namespace

## MCP HTTP protocol (how the server works)

The ark MCP client uses a simple HTTP transport - no SSE required:

```
POST /tools/list
  → {"tools": [{"name":"...","description":"...","inputSchema":{...}}, ...]}

POST /tools/call
  body: {"name": "list_incidents", "arguments": {}}
  → {"content": [{"type": "text", "text": "{\"incidents\":[...]}"}]}

POST /tools/call
  body: {"name": "get_incident_details", "arguments": {"id": "INC-001"}}
  → {"content": [{"type": "text", "text": "{\"id\":\"INC-001\",...}"}]}

GET /health  → 200
```

To adapt this server for a real backend: replace the `doList/doGetDetails/doEscalate`
functions with calls to your actual API (PagerDuty, Jira, OpsGenie, etc.) and add
auth headers via environment variables.

## Apply

```bash
# 1. Build the MCP server image (Docker Desktop shares the daemon - no push needed)
docker build -t incident-mcp:latest ./src

# 2. Deploy
kubectl apply -f secret.yaml          # add your Anthropic key first
kubectl apply -f incident-mcp.yaml
kubectl wait --for=condition=ready pod -l app=incident-mcp -n ark-teams --timeout=30s
kubectl apply -f incident-team.yaml
```

## Trigger

```bash
# Triage all incidents
ark trigger incident-triage-team -n ark-teams \
  --input '{"severity_filter":"all"}'

# Critical incidents only
ark trigger incident-triage-team -n ark-teams \
  --input '{"severity_filter":"critical"}'
```

## Expected output

```
RUNBOOK
=======
Most critical incident: INC-001 — API gateway 5xx spike (47 min open)

Likely root causes:
  1. Upstream dependency timeout / connection pool exhaustion
  ...

ESCALATION SUMMARY
==================
Escalated INC-001: severity=critical, open 47 minutes, no prior escalation.
No other incidents met the escalation criteria.
```

## Extending to a real MCP server

The only changes needed to go from mock to production:

1. Replace the `incidents` slice with calls to your real API
2. Add an env var for credentials (`MCP_API_TOKEN`) and read it in `main()`
3. Update `incident-mcp.yaml` to inject the secret

Everything else - the MCP transport, session management, tool dispatch - stays
identical. The `src/main.go` is designed to be forked.
