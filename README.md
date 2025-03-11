# Kubernetes Copilot

Kubernetes Copilot powered by LLM, which leverages advanced language models to streamline and enhance Kubernetes cluster management. This tool integrates seamlessly with your existing Kubernetes setup, providing intelligent automation, diagnostics, and manifest generation capabilities. By utilizing the power of AI, Kubernetes Copilot simplifies complex operations and helps maintain the health and security of your Kubernetes workloads.

## Features

- Automate Kubernetes cluster operations using ChatGPT (GPT-4 or GPT-3.5).
- Diagnose and analyze potential issues for Kubernetes workloads.
- Generate Kubernetes manifests based on provided prompt instructions.
- Utilize native `kubectl` and `trivy` commands for Kubernetes cluster access and security vulnerability scanning.
- Access the web and perform Google searches without leaving the terminal.

## Installation

Install the kube-copilot CLI with the following command:

```sh
go install github.com/feiskyer/kube-copilot/cmd/kube-copilot@latest
```

## Quick Start

Setup the following environment variables:

- Ensure [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/) is installed on the local machine and the kubeconfig file is configured for Kubernetes cluster access.
- Install [`trivy`](https://github.com/aquasecurity/trivy) to assess container image security issues (only required for the `audit` command).
- Set the OpenAI [API key](https://platform.openai.com/account/api-keys) as the `OPENAI_API_KEY` environment variable to enable ChatGPT functionality.

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
  -x, --max-iterations int   Max iterations for the agent running (default 10)
  -t, --max-tokens int       Max tokens for the GPT model (default 2048)
  -m, --model string         OpenAI model to use (default "gpt-4")
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

<summary>Azure OpenAI</summary>

For [Azure OpenAI service](https://learn.microsoft.com/en-us/azure/cognitive-services/openai/quickstart?tabs=command-line&pivots=rest-api#retrieve-key-and-endpoint), set the following environment variables:

- `AZURE_OPENAI_API_KEY=<your-api-key>`
- `AZURE_OPENAI_API_BASE=https://<replace-this>.openai.azure.com/`
- `AZURE_OPENAI_API_VERSION=2025-02-01-preview`

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
      --name string        Resource name
  -n, --namespace string   Resource namespace (default "default")
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
      --name string        Pod name
  -n, --namespace string   Pod namespace (default "default")

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
      --name string        Pod name
  -n, --namespace string   Pod namespace (default "default")

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
      --instructions string   instructions to execute

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

## Integrations

<details>
<summary>Google Search</summary>

Large language models are trained with outdated data, and hence may lack the most current information or miss out on recent developments. This is where Google Search becomes an optional tool. By integrating real-time search capabilities, LLMs can access the latest data, ensuring that responses are not only accurate but also up-to-date.

To enable it, set `GOOGLE_API_KEY` and `GOOGLE_CSE_ID` (obtain API key from [Google Cloud](https://cloud.google.com/docs/authentication/api-keys?visit_id=638154888929258210-4085587461) and CSE ID from [Google CSE](http://www.google.com/cse/)).
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
