# Kubernetes Copilot

Kubernetes Copilot powered by LLM, which leverages advanced language models to streamline and enhance Kubernetes cluster management. This tool integrates seamlessly with your existing Kubernetes setup, providing intelligent automation, diagnostics, and manifest generation capabilities. By utilizing the power of AI, Kubernetes Copilot simplifies complex operations and helps maintain the health and security of your Kubernetes workloads.

## Features

- Automate Kubernetes cluster operations using large language models.
- Provide your own OpenAI, Azure OpenAI, Anthropic Claude, Google Gemini or other OpenAI-compatible LLM providers.
- Diagnose and analyze potential issues for Kubernetes workloads.
- Generate Kubernetes manifests based on provided prompt instructions.
- Utilize native `kubectl` and `trivy` commands for Kubernetes cluster access and security vulnerability scanning.
- Support for Model Context Protocol (MCP) protocol to integrate with external tools.

## Installation

Install the kube-copilot CLI with the following command:

```sh
go install github.com/feiskyer/kube-copilot/cmd/kube-copilot@latest
```

## Quick Start

Setup the following environment variables:

- Ensure [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/) is installed on the local machine and the kubeconfig file is configured for Kubernetes cluster access.
- Install [`trivy`](https://github.com/aquasecurity/trivy) to assess container image security issues (only required for the `audit` command).
- Set the OpenAI [API key](https://platform.openai.com/account/api-keys) as the `OPENAI_API_KEY` environment variable to enable LLM AI functionality (refer below for other LLM providers).

Then run the following commands directly in the terminal:

```sh
Kubernetes Copilot powered by OpenAI

Usage:
  kube-copilot [command]

Available Commands:
  analyze     Analyze issues for a given resource
  audit       Audit security issues for a Pod
  completion  Generate the autocompletion script for the specified shell
  diagnose    Diagnose problems for a Pod
  execute     Execute operations based on prompt instructions
  generate    Generate Kubernetes manifests
  help        Help about any command
  version     Print the version of kube-copilot

Flags:
  -c, --count-tokens         Print tokens count
  -h, --help                 help for kube-copilot
  -x, --max-iterations int   Max iterations for the agent running (default 30)
  -t, --max-tokens int       Max tokens for the GPT model (default 2048)
  -m, --model string         OpenAI model to use (default "gpt-4o")
  -p, --mcp-config string    MCP configuration file
  -v, --verbose              Enable verbose output
      --version              version for kube-copilot

Use "kube-copilot [command] --help" for more information about a command.
```

## LLM Integrations

<details>
<summary>OpenAI</summary>

Set the OpenAI [API key](https://platform.openai.com/account/api-keys) as the `OPENAI_API_KEY` environment variable to enable OpenAI functionality.
</details>

<details>

<details>
<summary>Anthropic Claude</summary>

Anthropic Claude provides an [OpenAI compatible API](https://docs.anthropic.com/en/api/openai-sdk), so it could be used by using following config:

- `OPENAI_API_KEY=<your-anthropic-key>`
- `OPENAI_API_BASE='https://api.anthropic.com/v1/'`

</details>

<summary>Azure OpenAI</summary>

For [Azure OpenAI service](https://learn.microsoft.com/en-us/azure/cognitive-services/openai/quickstart?tabs=command-line&pivots=rest-api#retrieve-key-and-endpoint), set the following environment variables:

- `AZURE_OPENAI_API_KEY=<your-api-key>`
- `AZURE_OPENAI_API_BASE=https://<replace-this>.openai.azure.com/`
- `AZURE_OPENAI_API_VERSION=2025-03-01-preview`

</details>

<details>
<summary>Google Gemini</summary>

Google Gemini provides an OpenAI compatible API, so it could be used by using following config:

- `OPENAI_API_KEY=<your-google-ai-key>`
- `OPENAI_API_BASE='https://generativelanguage.googleapis.com/v1beta/openai/'`

</details>

<details>
<summary>Ollama or other OpenAI compatible LLMs</summary>

For Ollama or other OpenAI compatible LLMs, set the following environment variables:

- `OPENAI_API_KEY=<your-api-key>`
- `OPENAI_API_BASE='http://localhost:11434/v1'` (or your own base URL)

</details>

## Key Features

<details>
<summary>Analyze issues for a given kubernetes resource</summary>

`kube-copilot analyze [--resource pod] --name <resource-name> [--namespace <namespace>]` will analyze potential issues for the given resource object:

```sh
Analyze issues for a given resource

Usage:
  kube-copilot analyze [flags]

Flags:
  -h, --help               help for analyze
  -n, --name string        Resource name
  -s, --namespace string   Resource namespace (default "default")
  -r, --resource string    Resource type (default "pod")

Global Flags:
  -c, --count-tokens         Print tokens count
  -x, --max-iterations int   Max iterations for the agent running (default 10)
  -t, --max-tokens int       Max tokens for the GPT model (default 2048)
  -m, --model string         OpenAI model to use (default "gpt-4o")
  -v, --verbose              Enable verbose output
```

</details>

<details>
<summary>Audit Security Issues for Pod</summary>

`kube-copilot audit --name <pod-name> [--namespace <namespace>]` will audit security issues for a Pod:

```sh
Audit security issues for a Pod

Usage:
  kube-copilot audit [flags]

Flags:
  -h, --help               help for audit
  -n, --name string        Resource name
  -s, --namespace string   Resource namespace (default "default")

Global Flags:
  -c, --count-tokens         Print tokens count
  -x, --max-iterations int   Max iterations for the agent running (default 10)
  -t, --max-tokens int       Max tokens for the GPT model (default 2048)
  -m, --model string         OpenAI model to use (default "gpt-4o")
  -v, --verbose              Enable verbose output
```

</details>

<details>
<summary>Diagnose Problems for Pod</summary>

`kube-copilot diagnose --name <pod-name> [--namespace <namespace>]` will diagnose problems for a Pod:

```sh
Diagnose problems for a Pod

Usage:
  kube-copilot diagnose [flags]

Flags:
  -h, --help               help for diagnose
  -n, --name string        Resource name
  -s, --namespace string   Resource namespace (default "default")
  -p, --mcp-config string     MCP configuration file

Global Flags:
  -c, --count-tokens         Print tokens count
  -x, --max-iterations int   Max iterations for the agent running (default 10)
  -t, --max-tokens int       Max tokens for the GPT model (default 2048)
  -m, --model string         OpenAI model to use (default "gpt-4o")
  -v, --verbose              Enable verbose output
```

</details>

<details>
<summary>Execute operations based on prompt instructions</summary>

`kube-copilot execute --instructions <instructions>` will execute operations based on prompt instructions.
It could also be used to ask any questions.

```sh
Execute operations based on prompt instructions

Usage:
  kube-copilot execute [flags]

Flags:
  -h, --help                  help for execute
  -i, --instructions string   instructions to execute
  -p, --mcp-config string     MCP configuration file

Global Flags:
  -c, --count-tokens         Print tokens count
  -x, --max-iterations int   Max iterations for the agent running (default 10)
  -t, --max-tokens int       Max tokens for the GPT model (default 2048)
  -m, --model string         OpenAI model to use (default "gpt-4o")
  -v, --verbose              Enable verbose output
```

</details>

<details>
<summary>Generate Kubernetes Manifests</summary>

Use the `kube-copilot generate --prompt <prompt>` command to create Kubernetes manifests based on
the provided prompt instructions. After generating the manifests, you will be
prompted to confirm whether you want to apply them.

```sh
Generate Kubernetes manifests

Usage:
  kube-copilot generate [flags]

Flags:
  -h, --help            help for generate
  -p, --prompt string   Prompts to generate Kubernetes manifests

Global Flags:
  -c, --count-tokens         Print tokens count
  -x, --max-iterations int   Max iterations for the agent running (default 10)
  -t, --max-tokens int       Max tokens for the GPT model (default 2048)
  -m, --model string         OpenAI model to use (default "gpt-4o")
  -v, --verbose              Enable verbose output
```

</details>

<details>
<summary>Leverage Model Context Protocol (MCP)</summary>

Kube-copilot integrates with external tools for issue diagnosis and instruction execution (via the diagnose and execute subcommands) using the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/).

To use MCP tools:

1. Create a JSON configuration file for your MCP servers:

  ```json
  {
    "mcpServers": {
      "sequential-thinking": {
        "command": "npx",
        "args": [
          "-y",
          "@modelcontextprotocol/server-sequential-thinking"
        ]
      },
      "kubernetes": {
        "command": "uvx",
        "args": [
          "mcp-kubernetes-server"
        ]
      }
    }
  }
  ```

2. Run kube-copilot with the `--mcp-config` flag:

  ```sh
  kube-copilot execute --instructions "Your instructions" --mcp-config path/to/mcp-config.json
  ```

The MCP tools will be automatically discovered and made available to the LLM.

</details>

## Python Version

Please refer [feiskyer/kube-copilot-python](https://github.com/feiskyer/kube-copilot-python) for the Python implementation of the same project.

## Contribution

The project is opensource at github [feiskyer/kube-copilot](https://github.com/feiskyer/kube-copilot) (Go) and [feiskyer/kube-copilot-python](https://github.com/feiskyer/kube-copilot-python) (Python) with Apache License.

If you would like to contribute to the project, please follow these guidelines:

1. Fork the repository and clone it to your local machine.
2. Create a new branch for your changes.
3. Make your changes and commit them with a descriptive commit message.
4. Push your changes to your forked repository.
5. Open a pull request to the main repository.
