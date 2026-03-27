# 12 - Self-Optimizing Pipeline

**Pattern:** A pipeline that automatically improves its own prompts by analyzing run history, A/B testing a proposed revision, and (optionally) promoting it without human intervention.

**Use this when:** You have a pipeline running on a schedule or receiving steady traffic, and you want the system to detect and fix poorly-performing prompts automatically rather than waiting for a human to notice the failure rate.

## What it does

1. A two-step research pipeline (researcher → summarizer) accumulates run history.
2. Once 15 runs complete, `ArkOptimizer` computes per-step success rates.
3. If any step's success rate is below 95%, the optimizer dispatches a **meta-agent** (another Claude call) to propose a revised prompt based on the observed failure messages.
4. A **shadow team** is created with the proposed prompt and receives 15% of live traffic.
5. After 8 shadow runs, if the shadow's success rate improves by ≥15 percentage points over the original, the proposal is marked **pending** (or auto-promoted with `promotionPolicy: auto`).
6. A human approves or discards via `ark optimize approve/discard`.

## Resources created

| Resource | Purpose |
|---|---|
| `ArkTeam/research-team` | The original research pipeline |
| `ArkService/research-service` | Routes external traffic; patched to weighted during A/B tests |
| `ArkOptimizer/research-team-optimizer` | Watches run history and drives the improvement loop |

## Run it

```bash
# Deploy
kubectl apply -f research-team.yaml

# Trigger runs to build history (need ≥15 completed runs)
for i in $(seq 1 20); do
  ark trigger research-team -n ark-teams \
    --input "{\"topic\": \"Kubernetes operator pattern $i\"}"
done

# Check optimizer status
ark optimize status research-team -n ark-teams

# When a proposal is pending, review and approve (or discard)
ark optimize approve research-team -n ark-teams --reason "shadow win rate looks statistically solid"

# View full proposal history
ark optimize history research-team -n ark-teams
```

## Example optimizer status output

```
Optimizer:     research-team-optimizer
Target team:   research-team
Phase:         Pending
Observed runs: 0 / 15 required

STEP        RUNS  SUCCEEDED  FAILED  SUCCESS RATE  VAL FAILURES  AVG TOKENS
researcher  15    10         5       0.67          7             1842
summarizer  15    14         1       0.93          1             412

Active proposal: researcher-2026-03-24T09:14:02Z
  Step:          researcher
  Decision:      pending
  Shadow runs:   8 / 8
  Shadow rate:   0.88
  Original rate: 0.67
  Shadow team:   research-team-shadow-researcher-20260324t091402z

Proposed prompt:
You are a research agent. For every topic you investigate:
1. Search for at least 3 independent sources.
2. Cite each source with its URL in your response.
...

Reason:
The original prompt did not require source citation...
```

## Controlling the optimizer

```bash
# Pause (stops new cycles; current A/B test completes)
ark optimize pause research-team -n ark-teams

# Resume
ark optimize resume research-team -n ark-teams

# Use a different optimizer name
ark optimize status research-team -n ark-teams --optimizer my-custom-optimizer-name
```

## promotionPolicy: auto

To remove the human approval step and promote winning prompts automatically:

```yaml
spec:
  promotionPolicy: auto
```

With `auto`, when the shadow wins, the optimizer directly patches `ArkTeam.spec.pipeline[researcher].inputs.prompt` and the shadow team is deleted. The change is recorded in `status.history`.

## Tuning

| Field | Default | When to change |
|---|---|---|
| `minRuns` | 10 | Increase for noisier pipelines (more data = less false positives) |
| `shadowTrafficPct` | 10 | Increase for faster evaluation; keep ≤50 (shadow never majority) |
| `evaluationRuns` | 5 | Increase for higher statistical confidence |
| `minSuccessRateDelta` | 0.10 | Increase to require a larger improvement before promoting |
| `optimizerBudget.maxCyclesPerDay` | 3 | Decrease if meta-agent costs are a concern |
