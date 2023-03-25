# Kubernetes Copilot

ChatGPT based copilot for your Kubernetes cluster.

**Caution: Copilot may generate and execute inappropriate operations, do not use in production environment!**

## Install

Install the copilot with pip command below:

```sh
pip install kube-copilot
```

`kubectl` command should be installed in the local machine and kubeconfig file should be configured to access kubernetes cluster.

## How to use

```sh
Usage: kube-copilot [OPTIONS] COMMAND [ARGS]...

Options:
  --version  Show the version and exit.
  --short    Disable verbose information of copilot execution steps
  --help     Show this message and exit.

Commands:
  audit     audit security issues for a Pod
  diagnose  diagnose problems for a Pod
  execute   execute operations based on prompt instructions
```

## Contribution

The project is opensourced at github [feiskyer/kube-copilot](https://github.com/feiskyer/kube-copilot) with Apache License.

If you would like to contribute to the project, please follow these guidelines:

1. Fork the repository and clone it to your local machine.
2. Create a new branch for your changes.
3. Make your changes and commit them with a descriptive commit message.
4. Push your changes to your forked repository.
5. Open a pull request to the main repository.
