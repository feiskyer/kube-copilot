# Kubernetes Copilot

Kubernetes Copilot powered by OpenAI.

Features:

- Automate Kubernetes cluster operations using ChatGPT (GPT-4 or GPT-3.5).
- Diagnose and analyze the potential issues for Kubernetes workloads.
- Generate the Kubernetes manifests based on the provided prompt instructions.
- Utilize native kubectl and trivy commands for Kubernetes cluster access and security vulnerability scanning.
- Access the web and perform Google searches without leaving the terminal.

## Install

Install the copilot with the commands below:

```sh
go install github.com/feiskyer/kube-copilot/cmd/kube-copilot
```

## How to use

Setup the following environment variables:

- Ensure [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/) is installed on the local machine and the kubeconfig file is configured for Kubernetes cluster access.
- Install [`trivy`](https://github.com/aquasecurity/trivy) to assess container image security issues (only required for the `audit` command).
- Set the OpenAI [API key](https://platform.openai.com/account/api-keys) as the `OPENAI_API_KEY` environment variable to enable ChatGPT functionality.
  - For [Azure OpenAI service](https://learn.microsoft.com/en-us/azure/cognitive-services/openai/quickstart?tabs=command-line&pivots=rest-api#retrieve-key-and-endpoint), also set `OPENAI_API_TYPE=azure` and `OPENAI_API_BASE=https://<replace-this>.openai.azure.com/`.
- Google search is disabled by default. To enable it, set `GOOGLE_API_KEY` and `GOOGLE_CSE_ID` (obtain from [here](https://cloud.google.com/docs/authentication/api-keys?visit_id=638154888929258210-4085587461) and [here](http://www.google.com/cse/)).

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

Flags:
  -c, --count-tokens     Print tokens count
  -h, --help             help for kube-copilot
  -t, --max-tokens int   Max tokens for the GPT model (default 1024)
  -m, --model string     OpenAI model to use (default "gpt-4")
  -v, --verbose          Enable verbose output (default true)

Use "kube-copilot [command] --help" for more information about a command.
```

## Python Version

Please note that the original project (version number < v0.5.0) is written in Python 3 and the codes are in [main](https://github.com/feiskyer/kube-copilot/tree/main) branch.

## Contribution

The project is opensource at github [feiskyer/kube-copilot](https://github.com/feiskyer/kube-copilot) with Apache License.

If you would like to contribute to the project, please follow these guidelines:

1. Fork the repository and clone it to your local machine.
2. Create a new branch for your changes.
3. Make your changes and commit them with a descriptive commit message.
4. Push your changes to your forked repository.
5. Open a pull request to the main repository.
