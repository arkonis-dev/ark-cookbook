# 02 - Dynamic Team: Customer Support

**Pattern:** Autonomous delegation - no pipeline, coordinator decides routing at runtime.

**Use this when:** You can't know the execution path ahead of time. The right agent depends on the content of the task.

## What it does

A coordinator receives incoming support tickets and delegates to `billing`, `technical`, or `escalation` specialists based on the ticket content. No fixed order - the LLM decides.

## Features

- Dynamic mode (`entry` + `canDelegate`, no `pipeline` block)
- `delegate()` built-in tool injected into the coordinator at runtime
- Per-role `replicas` (scale specialists independently)
- `ArkService` with `least-busy` routing
- `AGENT_TEAM_ROUTES` env var injected by operator (role → queue URL map)

## Run it

```bash
kubectl apply -f customer-support.yaml

# Submit a ticket via the entry ArkService
curl -X POST http://<entry-service-ip>:8081/task \
  -H "Content-Type: application/json" \
  -d '{"prompt": "My invoice shows a double charge for last month."}'
```
