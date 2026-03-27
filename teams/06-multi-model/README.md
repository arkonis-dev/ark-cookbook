# 06 - Multi-Model: Cost-Optimized Content Pipeline

**Pattern:** Right model for the right task - mix providers and model tiers in one pipeline.

**Use this when:** Not every step needs your most powerful (and expensive) model. Classification, formatting, and simple transforms can use cheap models; reserve Sonnet/GPT-4o for steps that genuinely need reasoning.

## What it does

A content production pipeline that maps model cost to task complexity:

| Step | Model | Why |
|---|---|---|
| classify | claude-haiku | Classification - fast, cheap, no reasoning needed |
| research | claude-sonnet | Multi-step reasoning over complex topics |
| outline | gpt-4o-mini | Structured JSON output - GPT-4o-mini excels here |
| write | claude-sonnet | Creative, high-quality prose |
| polish | claude-haiku | Light grammar/formatting pass - no reasoning needed |

The deep research step is skipped for short/medium content (`if:` condition), saving ~60% on cost for most runs.

## Features

- Different `model:` per role (claude-haiku, claude-sonnet, gpt-4o-mini)
- Provider auto-detection: `claude-*` → anthropic, `gpt-*` → openai
- `ArkSettings` for shared config (temperature, style guide, output rules)
- Conditional step (`if: '{{ eq .input.length "long" }}'`)

## Run it

```bash
kubectl apply -f cost-optimized-content.yaml

ark trigger content-pipeline-team -n ark-teams \
  --input '{
    "topic": "How KEDA autoscaling works with AI agent workloads",
    "format": "blog",
    "length": "long"
  }'
```
