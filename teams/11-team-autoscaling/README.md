# 11 - Team Autoscaling (native, no KEDA)

**Pattern:** Pipeline with demand-based scaling and scale-to-zero.

**Use this when:** You want agents to scale with pipeline activity without installing KEDA.

## What it does

A 3-step pipeline (researcher → writer → editor) where all agents start at 0 replicas when idle and scale up automatically when a run is triggered.

## How autoscaling works

The operator computes desired replicas for each inline role on every reconcile:

```
desired = min(ceil(activeSteps / maxConcurrentTasks), role.replicas)
```

- `activeSteps` = number of Running/WarmingUp steps for this role across all concurrent runs
- `maxConcurrentTasks` = from `spec.roles[].limits.maxConcurrentTasks` (default 5)
- `role.replicas` = configured maximum (ceiling, never exceeded)

Scale-to-zero fires after `autoscaling.scaleToZero.afterSeconds` of inactivity.

## WarmingUp phase

When an agent is at 0 replicas and a step is ready to run, the step enters `WarmingUp` phase instead of `Running`. It stays there until the operator scales the agent up and pods become ready. No manual intervention needed - the transition to `Running` is automatic.

```
Pending → WarmingUp → Running → Succeeded
```

## Difference from KEDA autoscaling (03-cron-autoscale)

| | Team autoscaling (this) | KEDA autoscaling |
|-|------------------------|-----------------|
| External dependency | None | KEDA v2 in cluster |
| Scale trigger | Active pipeline steps | Redis queue depth |
| Best for | Pipeline-aware scaling | Queue-depth bursting |
| Config location | `ArkTeam.spec.autoscaling` | `ArkTeamRole.autoscaling` |

## Try it

```bash
kubectl apply -f autoscale-team.yaml -n ark-teams

# Trigger a run (agents start at 0 replicas)
ark trigger research-autoscale -n ark-teams \
  --input '{"topic":"neural network efficiency","audience":"ML engineers"}'

# Watch scale-up events
kubectl get events -n ark-teams --field-selector reason=AutoscaleAgent -w

# Check WarmingUp step state
kubectl get arkruns -n ark-teams

# After 5 minutes idle, agents scale back to 0
kubectl get arkagents -n ark-teams -o wide
```

## kubectl get arkagents -o wide output

When autoscaling is active, the `DESIRED` column shows the operator-computed target:

```
NAME                          MODEL        REPLICAS   READY   DESIRED   TOKENS(24H)   AGE
research-autoscale-researcher qwen2.5:7b   2          0       0         -             5m
research-autoscale-writer     qwen2.5:7b   2          0       0         -             5m
research-autoscale-editor     qwen2.5:7b   1          0       0         -             5m
```

After triggering:
```
NAME                          MODEL        REPLICAS   READY   DESIRED   TOKENS(24H)   AGE
research-autoscale-researcher qwen2.5:7b   1          1       1         -             6m
research-autoscale-writer     qwen2.5:7b   0          0       0         -             6m
research-autoscale-editor     qwen2.5:7b   0          0       0         -             6m
```
