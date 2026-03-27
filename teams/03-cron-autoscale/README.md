# 03 - Cron + Autoscale: Nightly Market Report

**Pattern:** Scheduled execution with KEDA autoscaling and token budget enforcement.

**Use this when:** Work arrives in bursts on a schedule. You want zero idle cost but fast scale-up when the queue fills.

## What it does

A cron event fires every night. Worker agents start at 0 replicas and scale up automatically as tasks accumulate (via KEDA). Each agent enforces a daily token budget. Agents scale back to 0 when the queue drains or the budget is exhausted.

## Features

- `ArkEvent` with `cron` source
- `spec.autoscaling` (KEDA `ScaledObject`) with `minReplicas: 0`
- `targetPendingTasks` as the scale trigger
- `maxDailyTokens` (rolling 24h budget → agent rejects tasks + operator scales to 0)
- `concurrencyPolicy: Forbid` (skip if previous run still going)
- Conditional step (`if:`) for optional forecast

## Prerequisites

KEDA must be installed: https://keda.sh/docs/deploy/

## Run it

```bash
kubectl apply -f nightly-report.yaml

# Trigger manually (no need to wait for midnight)
ark trigger nightly-report-team -n ark-teams \
  --input '{"reportDate": "2026-03-20", "includeForecasts": "true"}'

# Watch KEDA scale workers up as tasks queue
kubectl get pods -n ark-teams -w
```
