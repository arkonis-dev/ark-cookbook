<p align="center">
  <img src="https://raw.githubusercontent.com/arkonis-dev/ark-cookbook/main/assets/logo.svg" width="120" alt="ark">
</p>

<h1 align="center">ark-cookbook</h1>

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](./LICENSE)

Community recipes for [ark](https://github.com/arkonis-dev/ark) - production-inspired AI agent pipelines as Kubernetes-native YAML.

Each recipe is a self-contained YAML file that highlights a specific pattern. Read the header comments - they explain what's demonstrated and how to run it.

Full documentation at **[docs.arkonis.dev](https://docs.arkonis.dev)**.

## Recipes

| Recipe | Pattern | Key Features |
|---|---|---|
| [01-simple-pipeline](teams/01-simple-pipeline/) | Linear DAG | `pipeline`, `dependsOn`, template inputs/outputs, `outputSchema` |
| [02-dynamic-team](teams/02-dynamic-team/) | Autonomous delegation | `entry`, `canDelegate`, `delegate()` tool, `ArkService` |
| [03-cron-autoscale](teams/03-cron-autoscale/) | Cron + scale-to-zero | `ArkEvent` cron, KEDA `autoscaling`, `maxDailyTokens` |
| [04-webhook-pipeline](teams/04-webhook-pipeline/) | Webhook fan-out | `ArkEvent` webhook, fan-out to multiple teams, inline `tools` |
| [05-validated-pipeline](teams/05-validated-pipeline/) | Output validation + retry | `validate` (contains/schema/semantic), `onFailure: retry`, `loop` |
| [06-multi-model](teams/06-multi-model/) | Cost-optimized pipeline | Mixed models per role, `ArkSettings`, conditional steps |
| [07-supervisor-worker](teams/07-supervisor-worker/) | Horizontal scale | `submit_subtask`, `replicas`, `maxConcurrentTasks`, `least-busy` routing |
| [08-budget-notify](teams/08-budget-notify/) | Production ops | `maxDailyTokens`, `ArkNotify`, `notifyRef`, `systemPromptRef` |

## Prerequisites

- Kubernetes cluster
- [ark-cli](https://github.com/arkonis-dev/ark-cli) installed
- ark installed:
  ```bash
  helm repo add arkonis https://arkonis-dev.github.io/helm-charts
  helm repo update
  helm install ark arkonis/ark --namespace ark-system --create-namespace
  kubectl create secret generic my-ark-secrets \
    --namespace ark-system \
    --from-literal=ANTHROPIC_API_KEY=sk-ant-...
  helm upgrade ark arkonis/ark \
    --namespace ark-system \
    --set apiKeys.existingSecret=my-ark-secrets
  ```

For recipe 03 (autoscaling): KEDA must be installed - see [keda.sh/docs/deploy](https://keda.sh/docs/deploy/).

## Quick start

```bash
# Apply a recipe
kubectl apply -f teams/01-simple-pipeline/blog-writer.yaml

# Trigger a run
ark trigger blog-writer-team -n ark-teams \
  --input '{"topic": "Kubernetes operators explained", "audience": "platform engineers"}'

# Watch it run
ark status -n ark-teams
```

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for how to add a recipe.

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](./CODE_OF_CONDUCT.md).

## License

Apache 2.0 - see [LICENSE](LICENSE).
