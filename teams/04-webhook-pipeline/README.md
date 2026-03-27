# 04 - Webhook Pipeline: CI/CD Code Review

**Pattern:** External event triggers parallel pipelines (fan-out).

**Use this when:** An external system (GitHub, PagerDuty, Stripe, etc.) needs to kick off agent workflows via HTTP.

## What it does

A GitHub webhook fires on pull request open/update. One webhook event spawns two teams in parallel: a security scanner and a code quality reviewer. Both post their findings back to the PR as GitHub review comments.

## Features

- `ArkEvent` with `webhook` source (auto-generates endpoint + Bearer token)
- Fan-out: `targets` dispatches to multiple teams from one event
- `{{ .trigger.body.* }}` to access the POST payload in template inputs
- Inline `tools` (no MCP server - agent calls GitHub API directly)
- `concurrencyPolicy: Allow` (multiple PRs reviewed concurrently)

## Run it

```bash
kubectl apply -f ci-review.yaml

# Get the webhook URL and token
WEBHOOK_URL=$(kubectl get arkevent pr-review-webhook -o jsonpath='{.status.webhookURL}')
TOKEN=$(kubectl get secret pr-review-webhook-webhook-token \
          -o jsonpath='{.data.token}' | base64 -d)

# Test with a simulated PR event
curl -X POST "$WEBHOOK_URL" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "pr_number": 42,
    "repo": "my-org/my-service",
    "branch": "feat/new-auth",
    "diff_url": "https://github.com/my-org/my-service/pull/42.diff",
    "author": "alice"
  }'
```

Configure in GitHub: Settings → Webhooks → Add webhook → paste `$WEBHOOK_URL` with header `Authorization: Bearer $TOKEN`.
