# 05 - Routed Team: Ops Router

**Pattern:** ArkTeam routed mode — LLM-driven capability dispatch (RFC-0019).

**Use this when:** You have a pool of specialist agents and want to route incoming tasks
to the right one automatically, without hard-coding which agent handles what in your
pipeline YAML or event sources.

## What it does

An `ops-router` team receives free-form engineering tasks and dispatches each one to the
most appropriate specialist — without any pipeline or roles defined. The operator makes a
single lightweight LLM call (the "router") that reads the indexed agent capabilities and
selects the best match, then submits the task directly to that agent's queue.

| Agent | Capabilities | Handles |
|---|---|---|
| `frontend-engineer` | `frontend-debug`, `frontend-performance` | React/TypeScript bugs, CSS, browser quirks, Web Vitals |
| `backend-engineer` | `backend-debug`, `database-query` | API errors, slow queries, N+1, schema issues |
| `devops-engineer` | `incident-response`, `deployment-debug`, `cloud-cost` | Production incidents, K8s rollouts, FinOps |
| `security-auditor` | `security-review` | OWASP vulnerabilities, IAM, secrets, infra audits |
| `general-assistant` | _(fallback — no capability declared)_ | Cross-domain questions, standup summaries |

The router itself uses `qwen2.5:7b` — cheap and fast for the dispatch call. Each specialist
uses the same model for the actual work; swap to a larger model per-agent as needed.

## How routed mode differs from recipe 10 (registry-routing)

| | Recipe 10 | Recipe 14 |
|---|---|---|
| RFC | RFC-0001 | RFC-0019 |
| Feature | `registryLookup` on pipeline steps | `spec.routing` on ArkTeam |
| Has pipeline? | Yes | No |
| Has roles? | Yes | No |
| Who routes? | Operator resolves one step's agent | Router LLM dispatches the whole task |
| When to use | Some steps need dynamic agents; others are fixed | The whole task goes to one of N specialists |

## Features

- `spec.routing` — routed mode; no pipeline or roles
- `spec.routing.model` — lightweight model for the router call
- `spec.routing.fallback` — named agent used when no capability matches (or for cross-domain tasks)
- `ArkRegistry` — auto-indexes all agents in the namespace by capability
- Multi-capability agent (`devops-engineer` handles three capability IDs)
- `ArkEvent` cron — routes a daily standup summary automatically

## Run it

```bash
kubectl apply -f secret.yaml
kubectl apply -f ops-router.yaml

# Route a frontend bug
ark trigger ops-router -n ark-teams \
  --input '{"task": "The login button is broken on Safari iOS 17"}'

# Route a database issue
ark trigger ops-router -n ark-teams \
  --input '{"task": "Our PostgreSQL query for user search is taking 8 seconds"}'

# Route a production incident
ark trigger ops-router -n ark-teams \
  --input '{"task": "CPU on api-gateway spiked to 95% — pods are crashing"}'

# Route a security review
ark trigger ops-router -n ark-teams \
  --input '{"task": "Review this auth middleware for security issues", "code": "..."}'

# Cross-domain standup summary (routed to general-assistant fallback)
ark trigger ops-router -n ark-teams \
  --input '{"task": "Summarize key health signals for standup: incidents, costs, security findings"}'
```

Check the routing decision (which agent was selected and why):

```bash
kubectl get arkrun -n ark-teams -l arkonis.dev/team=ops-router --sort-by=.metadata.creationTimestamp
kubectl describe arkrun <run-name> -n ark-teams
```

The `status.steps[0]` of each run shows:

```yaml
steps:
  - name: route
    phase: Succeeded
    resolvedAgent: frontend-engineer
    selectedCapability: frontend-debug
    routingReason: "Task describes a UI bug on a specific browser — matches frontend debugging."
    output: "..."
```

## What the routing decision looks like

The router receives all indexed capabilities as a structured list and the task input,
then responds with JSON:

```json
{"capability": "frontend-debug", "reason": "Task describes a UI bug on Safari iOS 17."}
```

The operator resolves `frontend-debug` → `frontend-engineer` via ArkRegistry, submits
the task to that agent's queue, and records the routing decision in the ArkRun status.

## Cron auto-routing

The `daily-ops-summary` ArkEvent fires at 08:00 UTC on weekdays. The standup summary
task is cross-domain (incidents + costs + security), so the router falls back to
`general-assistant`, which has a broad system prompt designed for exactly this synthesis.

Routed mode dispatches to **one agent per trigger**, but that agent can answer any question.
A cross-domain standup summary is a perfectly valid routed task: the selected agent
(or the fallback) produces the full multi-domain output in a single response. Use a
**pipeline** only when you want multiple specialists to each contribute their domain
independently and have their outputs composed.
