<p align="center">
  <img src="https://raw.githubusercontent.com/arkonis-dev/ark-cookbook/main/assets/logo.svg" width="120" alt="ark">
</p>

<h1 align="center">ark-cookbook</h1>

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](./LICENSE)

Community recipes for [ark](https://github.com/arkonis-dev/ark) - production-inspired AI agent pipelines as Kubernetes-native YAML.

Each recipe is a self-contained YAML file that highlights a specific pattern. Read the header comments - they explain what's demonstrated and how to run it.

Full documentation at **[docs.arkonis.dev](https://docs.arkonis.dev)**.

## Recipes

| Recipe                                              | Pattern               | Key Features                                                                         |
| --------------------------------------------------- | --------------------- | ------------------------------------------------------------------------------------ |
| [01-simple-pipeline](teams/01-simple-pipeline/)     | Linear DAG            | `pipeline`, `dependsOn`, template inputs/outputs, `outputSchema`                     |
| [02-dynamic-team](teams/02-dynamic-team/)           | Autonomous delegation | `entry`, `canDelegate`, `delegate()` tool, `ArkEvent` webhook                        |
| [03-budget-notify](teams/03-budget-notify/)         | Production ops        | `maxDailyTokens`, `ArkNotify`, `notifyRef`, `systemPromptRef`                        |
| [04-single-agent](teams/04-single-agent/)           | Minimal single agent  | One role, one step, `outputSchema`, simplest possible setup                          |
| [05-routed-team](teams/05-routed-team/)             | LLM-driven dispatch   | `spec.routing`, `ArkRegistry`, multi-capability agents, cron auto-routing            |
| [06-custom-mcp-server](teams/06-custom-mcp-server/) | Custom MCP server     | Build and deploy your own MCP tool server; read + write tools; `mcpServers[].url`    |
| [07-postgres-mcp](teams/07-postgres-mcp/)           | Postgres MCP          | Real database via custom MCP; parallel `analyst` + `auditor` steps; live SQL queries |

## Prerequisites

- Kubernetes cluster
- [ark-cli](https://github.com/arkonis-dev/ark-cli) installed
- ark installed:

```bash
helm repo add arkonis https://arkonis-dev.github.io/helm-charts/
helm repo update
helm install ark arkonis/ark \
  --namespace ark-system \
  --create-namespace
```

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
