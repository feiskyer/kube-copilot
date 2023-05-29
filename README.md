# Kubernetes Copilot

Kubernetes Copilot powered by OpenAI.

**Status: Experimental**

**Caution: Copilot may generate and execute inappropriate operations, do not use in production environment!**

Features:

* Automatically operate Kubernetes cluster based on prompt instructions.
* Human interactions on uncertain instructions to avoid inappropriate operations.
* Native kubectl and bash commands for accessing Kubernetes cluster.
* Web access and Google search support without leaving the terminal.

## Install

Install the copilot with pip command below:

```sh
pip install kube-copilot
```

## Setup

* `kubectl` should be [installed](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/) in the local machine and kubeconfig file should be configured to access kubernetes cluster.
* `trivy` should be [installed](https://github.com/aquasecurity/trivy) to evaluate the security issues of container images (for `audit` command).
* OpenAI [API key](https://platform.openai.com/account/api-keys) should be set to `OPENAI_API_KEY` environment variable to enable the ChatGPT feature.
  * `OPENAI_API_TYPE=azure` and `OPENAI_API_BASE=https://<replace-this>.openai.azure.com/` should be set as well for [Azure OpenAI service](https://learn.microsoft.com/en-us/azure/cognitive-services/openai/quickstart?tabs=command-line&pivots=rest-api#retrieve-key-and-endpoint).
* Google search is not enabled by default. Set `GOOGLE_API_KEY` and `GOOGLE_CSE_ID` if you want to enable it (get from [here](https://cloud.google.com/docs/authentication/api-keys?visit_id=638154888929258210-4085587461) and [here](http://www.google.com/cse/ )).

## How to use

```sh
Usage: kube-copilot [OPTIONS] COMMAND [ARGS]...

  Kubernetes Copilot powered by OpenAI

Options:
  --version  Show the version and exit.
  --help     Show this message and exit.

Commands:
  audit     audit security issues for a Pod
  diagnose  diagnose problems for a Pod
  execute   execute operations based on prompt instructions
```

### Audit Security Issues for Pod

`kube-copilot audit POD [NAMESPACE]` will audit security issues for a Pod:

```sh
Usage: kube-copilot audit [OPTIONS] POD [NAMESPACE]

  audit security issues for a Pod

Options:
  --verbose     Enable verbose information of copilot execution steps
  --model TEXT  OpenAI model to use for copilot execution, default is gpt-4
  --help        Show this message and exit.
```

### Diagnose Problems for Pod

`kube-copilot diagnose POD [NAMESPACE]` will diagnose problems for a Pod:

```sh
Usage: kube-copilot diagnose [OPTIONS] POD [NAMESPACE]

  diagnose problems for a Pod

Options:
  --verbose     Enable verbose information of copilot execution steps
  --model TEXT  OpenAI model to use for copilot execution, default is gpt-4
  --help        Show this message and exit.
```

### Execute Operations Based on Prompt Instructions

`kube-copilot execute INSTRUCTIONS` will execute operations based on prompt instructions.
It could also be used to ask any questions.

```sh
Usage: kube-copilot execute [OPTIONS] INSTRUCTIONS

  execute operations based on prompt instructions

Options:
  --verbose     Enable verbose information of copilot execution steps
  --model TEXT  OpenAI model to use for copilot execution, default is gpt-4
  --help        Show this message and exit.
```

## Contribution

The project is opensource at github [feiskyer/kube-copilot](https://github.com/feiskyer/kube-copilot) with Apache License.

If you would like to contribute to the project, please follow these guidelines:

1. Fork the repository and clone it to your local machine.
2. Create a new branch for your changes.
3. Make your changes and commit them with a descriptive commit message.
4. Push your changes to your forked repository.
5. Open a pull request to the main repository.
