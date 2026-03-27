# 08 - Budget + Notify: Analytics Team

**Pattern:** Production-grade ops - cost controls, Slack alerts, external prompt management.

**Use this when:** You're running in production and need visibility into cost, failures, and timeouts without watching dashboards.

## What it does

An analytics team that answers business questions by querying a data warehouse. Every lifecycle event (success, failure, timeout, budget hit) fires a Slack notification. Daily token budgets prevent runaway spend. The analyst's system prompt lives in a ConfigMap - update it without touching the ArkTeam spec.

## Features

- `maxDailyTokens` (rolling 24h token budget - agent rejects tasks + operator scales to 0)
- `ArkNotify` CRD with Slack channel
- `notifyRef` on `ArkTeam` (routes events to notify channel)
- All four event types: `TeamSucceeded`, `TeamFailed`, `TeamTimedOut`, `BudgetExceeded`
- Custom Slack message templates with `{{ .team.* }}`, `{{ .run.* }}`, `{{ .agent.* }}`
- Rate limiting (1 alert per event type per 5 minutes)
- `systemPromptRef.configMapKeyRef` (large prompt loaded from ConfigMap)
- Inline `tools` (SQL query + schema fetch, no MCP needed)
- Semantic output validation with retry

## Run it

```bash
# 1. Create the Slack webhook secret
kubectl create secret generic analytics-slack-secret \
  --from-literal=webhookUrl=https://hooks.slack.com/services/T.../B.../... \
  -n ark-teams

# 2. Create the system prompt ConfigMap
kubectl create configmap analytics-analyst-prompt \
  --from-literal=system.txt="You are a data analyst. Use the run_query tool to answer questions." \
  -n ark-teams

# 3. Apply
kubectl apply -f analytics-team.yaml

# 4. Trigger
ark trigger analytics-team -n ark-teams \
  --input '{
    "dataset": "user_events_2026_q1",
    "question": "What drove the 15% drop in activation rate in week 10?"
  }'
```
