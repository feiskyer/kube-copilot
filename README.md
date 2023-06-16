# Kubernetes Copilot

Kubernetes Copilot powered by OpenAI.

**Status:** Experimental.

**Caution:** Copilot may generate and execute unsuitable operations; avoid using in production environments! For improved reasoning performance, **GPT-4** is highly recommended (although GPT-3.5 is also supported).

Features:

- Automate Kubernetes cluster operations using Plan-and-Solve prompts with GPT-4.
- Utilize native kubectl and trivy commands for Kubernetes cluster access and security vulnerability scanning.
- Access the web and perform Google searches without leaving the terminal.

## Install

Install the copilot with pip command below:

```sh
pip install kube-copilot
```

## Setup

- Ensure [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/) is installed on the local machine and the kubeconfig file is configured for Kubernetes cluster access.
- Install [`trivy`](https://github.com/aquasecurity/trivy) to assess container image security issues (for the `audit` command).
- Set the OpenAI [API key](https://platform.openai.com/account/api-keys) as the `OPENAI_API_KEY` environment variable to enable ChatGPT functionality.
  - For [Azure OpenAI service](https://learn.microsoft.com/en-us/azure/cognitive-services/openai/quickstart?tabs=command-line&pivots=rest-api#retrieve-key-and-endpoint), also set `OPENAI_API_TYPE=azure` and `OPENAI_API_BASE=https://<replace-this>.openai.azure.com/`.
- Google search is disabled by default. To enable it, set `GOOGLE_API_KEY` and `GOOGLE_CSE_ID` (obtain from [here](https://cloud.google.com/docs/authentication/api-keys?visit_id=638154888929258210-4085587461) and [here](http://www.google.com/cse/)).

## How to use

Running directly in the terminal:

```sh
Usage: kube-copilot [OPTIONS] COMMAND [ARGS]...

  Kubernetes Copilot powered by OpenAI

Options:
  --version  Show the version and exit.
  --help     Show this message and exit.

Commands:
  analyze   analyze issues for a given resource
  audit     audit security issues for a Pod
  diagnose  diagnose problems for a Pod
  execute   execute operations based on prompt instructions
```

Running as container:

```sh
kubectl run -it --rm copilot \
  --env="OPENAI_API_KEY=$OPENAI_API_KEY" \
  --restart=Never \
  --image=ghcr.io/feiskyer/kube-copilot \
  -- execute --verbose 'What Pods are using max memory in the cluster'
```

### Audit Security Issues for Pod

`kube-copilot audit POD [NAMESPACE]` will audit security issues for a Pod:

```sh
Usage: kube-copilot audit [OPTIONS] POD [NAMESPACE]

  audit security issues for a Pod

Options:
  --verbose      Enable verbose information of copilot execution steps
  --model MODEL  OpenAI model to use for copilot execution, default is gpt-4
  --help         Show this message and exit.
```

### Diagnose Problems for Pod

`kube-copilot diagnose POD [NAMESPACE]` will diagnose problems for a Pod:

```sh
Usage: kube-copilot diagnose [OPTIONS] POD [NAMESPACE]

  diagnose problems for a Pod

Options:
  --verbose      Enable verbose information of copilot execution steps
  --model MODEL  OpenAI model to use for copilot execution, default is gpt-4
  --help         Show this message and exit.
```

### Analyze Potential Issues for k8s Object

`kube-copilot analyze RESOURCE NAME [NAMESPACE]` will analyze potential issues for the given resource object:


```sh
Usage: kube-copilot analyze [OPTIONS] RESOURCE NAME [NAMESPACE]

  analyze issues for a given resource

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
  --verbose      Enable verbose information of copilot execution steps
  --model MODEL  OpenAI model to use for copilot execution, default is gpt-4
  --help         Show this message and exit.
```

## Contribution

The project is opensource at github [feiskyer/kube-copilot](https://github.com/feiskyer/kube-copilot) with Apache License.

If you would like to contribute to the project, please follow these guidelines:

1. Fork the repository and clone it to your local machine.
2. Create a new branch for your changes.
3. Make your changes and commit them with a descriptive commit message.
4. Push your changes to your forked repository.
5. Open a pull request to the main repository.
