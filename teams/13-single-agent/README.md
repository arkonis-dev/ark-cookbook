# 13 - Single Agent: Daily Tech Brief

**Pattern:** Single agent - the simplest possible ark setup.

**Use this when:** You need one agent to handle a task end-to-end with no chaining.

## What it does

One analyst agent takes a topic and returns a structured tech brief - summary, trends, and talking points.

## Features

- Single role, single pipeline step
- `spec.inputs` with required/optional fields
- `outputSchema` (JSON Schema validation)
- Minimal configuration - good starting point for new users

## Run it

```bash
kubectl apply -f daily-tech-brief.yaml

ark trigger daily-tech-brief -n ark-teams \
  --input '{"topic": "AI agents in Kubernetes"}'
```

Try it locally without a cluster:

```bash
ark run daily-tech-brief.yaml \
  --input '{"topic": "AI agents in Kubernetes"}'
```

Mock mode:

```bash
ark run daily-tech-brief.yaml --provider mock \
  --input '{"topic": "AI agents in Kubernetes"}'
```

## Example output

```json
{
  "topic": "AI agents in Kubernetes",
  "summary": "Kubernetes is becoming the default runtime for AI agents, enabling scalable orchestration of LLM-based workloads alongside traditional services.",
  "trends": [
    "Operators like ark managing agent lifecycle as first-class resources",
    "MCP (Model Context Protocol) standardising tool access for agents",
    "Queue-based autoscaling replacing CPU metrics for LLM workloads"
  ],
  "talkingPoints": [
    "Agents scale like deployments - define replicas, the operator handles the rest",
    "Any model, any provider - Anthropic, OpenAI, or custom via plugin interface",
    "Full observability via OpenTelemetry - no vendor lock-in"
  ]
}
```
