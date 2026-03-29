# 02 - Dynamic Team: Customer Support

**Pattern:** Autonomous delegation - no pipeline, coordinator decides routing at runtime.

**Use this when:** You can't know the execution path ahead of time. The right agent depends on the content of the task.

## What it does

A coordinator receives incoming support tickets and delegates to `billing`, `technical`, or `escalation` specialists based on the ticket content. No fixed order - the LLM decides.

## Features

- Dynamic mode (`entry` + `canDelegate`, no `pipeline` block)
- `delegate()` built-in tool injected into the coordinator at runtime
- Per-role `replicas` (scale specialists independently)
- `ArkEvent` webhook for external task submission

## Run it

```bash
kubectl apply -f customer-support.yaml

# Get the webhook URL
kubectl get arkevent customer-support-webhook -n ark-teams \
  -o jsonpath='{.status.webhookURL}'

# Submit a ticket
curl -X POST "$WEBHOOK_URL" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"message": "My invoice shows a double charge for last month."}'
```
