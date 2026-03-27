# 07 - Supervisor/Worker: Document Analysis

**Pattern:** Supervisor breaks work into chunks, workers process them concurrently.

**Use this when:** A single task is too large for one LLM call (e.g. a 100-page document). Split it, parallelize, then synthesize.

## What it does

A supervisor receives a full document, splits it into ~500-word sections, and dispatches each as an independent sub-task via `submit_subtask`. Five worker pods (each handling 3 sections concurrently) process 15 sections in parallel. The supervisor synthesizes all results into a final report.

## Features

- `submit_subtask` built-in tool (supervisor dispatches workers at runtime)
- `replicas: 5` on the worker role (concurrent processing pods)
- `maxConcurrentTasks: 3` (per-pod parallelism)
- KEDA autoscaling: workers scale from 2 to 20 based on queue depth
- `ArkService` with `least-busy` routing (tasks go to least loaded pod)
- Dynamic mode: supervisor decides the split at runtime

## Concurrency math

```
5 pods × 3 concurrent tasks = 15 sections processed simultaneously
With KEDA: up to 20 pods × 3 = 60 sections simultaneously on burst
```

## Run it

```bash
kubectl apply -f doc-analysis.yaml

# Submit a document via the entry ArkService
curl -X POST http://<entry-service-ip>:8081/task \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Analyze this contract: <paste document text here>"}'
```
