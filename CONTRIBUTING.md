# Contributing to ark-cookbook

Thank you for your interest in contributing. The cookbook is a collection of example pipelines for [ark](https://github.com/arkonis-dev/ark). Contributions are YAML-only - no Go code required.

## Before you start

- Check existing recipes under `teams/` to avoid duplicating a pattern that already exists.
- If you're unsure whether a recipe fits, open an issue first to discuss.

## Adding a recipe

Each recipe lives in its own numbered directory under `teams/`:

```
teams/
  NN-short-name/
    README.md            # What the recipe does, when to use it, how to run it
    team.yaml            # The ArkFlow / ArkAgent / ArkService manifests
    values.yaml          # Optional: Helm overrides if the recipe needs custom settings
```

**Naming convention** - use the next available number and a short kebab-case name (e.g., `14-parallel-research`).

### README.md structure

Every recipe README must include:

1. **What it does** - 2–3 sentences describing the pipeline.
2. **When to use it** - the problem this pattern solves.
3. **Prerequisites** - any external services, MCP servers, or API keys needed.
4. **How to run** - exact commands to apply the manifests and trigger the pipeline.
5. **How it works** - a brief walkthrough of the YAML structure.

### YAML requirements

- Use `ark run team.yaml --provider mock` to verify the pipeline runs locally before submitting.
- All `systemPrompt` values must describe a realistic task - avoid placeholder text like "You are a helpful assistant."
- Multi-step flows must include `dependsOn` where steps have dependencies.
- Include resource `limits` (`maxTokensPerCall`, `timeoutSeconds`) on all `ArkAgent` specs.

## Security practices

- **Redact all secrets** - any API key, token, or credential that appears in a YAML sample must be replaced with a placeholder (e.g., `sk-ant-...`, `your-webhook-token`). Never commit a real key.
- **No `latest` image tags** - use pinned version tags in any `image:` field.
- **Use `apiKeys.existingSecret`** in Helm-based recipes rather than inline key values.
- **No hardcoded URLs to internal systems** - use placeholder URLs like `https://your-mcp-server.example.com/sse` for MCP endpoints.

## Validating a recipe locally

```bash
# Validate YAML structure and DAG
ark validate teams/NN-short-name/team.yaml

# Run with mock provider (no API key needed)
ark run teams/NN-short-name/team.yaml --provider mock --watch
```

## Submitting a pull request

1. Fork the repo and create a branch from `main`.
2. Add your recipe directory with a complete `README.md` and `team.yaml`.
3. Verify it passes `ark validate` and runs with `--provider mock`.
4. Open a PR against `main`. Include sample output (from `--provider mock --watch`) in the PR description.

We use **Rebase and merge** to keep a linear history on `main`.

## Reporting issues with existing recipes

Open a [GitHub issue](https://github.com/arkonis-dev/ark-cookbook/issues/new) with:

- The recipe name and directory
- ark version
- The error or unexpected behavior observed

## License

By contributing, you agree that your contributions will be licensed under the [Apache 2.0 License](./LICENSE).
