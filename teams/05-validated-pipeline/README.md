# 05 - Validated Pipeline: Structured Data Extraction

**Pattern:** Pipeline with output contracts - every step validates its output before proceeding.

**Use this when:** Your pipeline produces structured data that downstream systems depend on. Garbage-in-garbage-out is unacceptable.

## What it does

Extracts a structured CRM record (contact, deal, or event) from unstructured text. Each step enforces output quality via validation - regex, JSON Schema, or a semantic LLM check. Failed steps retry automatically up to the configured limit.

## Validation modes demonstrated

| Step | Mode | What it checks |
|---|---|---|
| extract | `contains` | Output starts with `{` (is JSON) |
| extract | `schema` | Required fields are present |
| enrich | `schema` | All enrichment fields present |
| enrich | `semantic` | LLM verifies enrichment makes business sense |
| format | `contains` | Output contains a Markdown heading |

## Features

- `validate:` block with `contains`, `schema`, and `semantic` modes
- `onFailure: retry` with `maxRetries`
- `loop` step with `loop.condition` (retry until confidence ≥ 0.8)
- `validationAttempts` tracked in `ArkRun.status.steps[].validationAttempts`

## Run it

```bash
kubectl apply -f data-extraction.yaml

ark trigger data-extraction-team -n ark-teams \
  --input '{
    "rawText": "John Smith, CTO at Acme Corp. Reached out on 2026-03-15 about enterprise pricing. Budget: $50k/year.",
    "targetSchema": "contact"
  }'
```
