# 01 - Simple Pipeline: Blog Writer

**Pattern:** Linear DAG - steps execute in a fixed order, each feeding the next.

**Use this when:** You have a deterministic workflow where you know the steps upfront and outputs chain predictably.

## What it does

Researcher → Writer → Editor. Each step uses a different model sized to the task complexity.

## Features

- `pipeline` mode with `dependsOn`
- `spec.inputs` with required/optional fields and defaults
- Template expressions: `{{ .input.* }}` and `{{ .steps.<name>.output }}`
- `outputSchema` (JSON Schema validation on the researcher's output)
- `timeoutSeconds` at the team level

## Run it

```bash
kubectl apply -f blog-writer.yaml

ark trigger blog-writer-team -n ark-teams \
  --input '{"topic": "Kubernetes operators explained", "audience": "platform engineers"}'
```

Try it locally without a cluster:

```bash
ark run blog-writer.yaml --provider mock \
  --input '{"topic": "Kubernetes operators explained"}'
```
