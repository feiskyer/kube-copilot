# Kubernetes Copilot

Kubernetes Copilot powered by OpenAI.

**Caution: Copilot may generate and execute inappropriate operations, do not use in production environment!**

**Major features:**

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

* `kubectl` should be installed in the local machine and kubeconfig file should be configured to access kubernetes cluster.
* `docker` should be installed to evaluate the security issues of container images (for `audit` command).
* OpenAI API key should be set to `OPENAI_API_KEY` environment variable to enable the ChatGPT feature.
  * `OPENAI_API_BASE` should be set as well for Azure OpenAI service.

## How to use

```sh
Usage: kube-copilot [OPTIONS] COMMAND [ARGS]...

  Kubernetes Copilot powered by OpenAI

Options:
  --version     Show the version and exit.
  --short       Disable verbose information of copilot execution steps
  --model TEXT  OpenAI model to use for copilot execution, default is
                gpt-3.5-turbo
  --help        Show this message and exit.

Commands:
  audit     audit security issues for a Pod
  diagnose  diagnose problems for a Pod
  execute   execute operations based on prompt instructions
```

### Audit Security Issues for Pod

`kube-copilot audit POD [NAMESPACE]` will audit security issues for a Pod.

Here is an example of auditing a Pod with old nginx image (note `image cve` is the ad-hoc prompt during the auditing process):

```sh
# kube-copilot audit nginx-6cc9f5fc9f-qxlxt
> Entering new AgentExecutor chain...
Thought: I need to gather more information about the Pod and its configuration to start the audit.

Action:
{
  "action": "Terminal",
  "action_input": "kubectl get pod nginx-6cc9f5fc9f-qxlxt -n default -o yaml"
}

Observation: apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2023-03-25T11:08:16Z"
  generateName: nginx-6cc9f5fc9f-
  labels:
    app: nginx
    pod-template-hash: 6cc9f5fc9f
  name: nginx-6cc9f5fc9f-qxlxt
  namespace: default
  ownerReferences:
  - apiVersion: apps/v1
    blockOwnerDeletion: true
    controller: true
    kind: ReplicaSet
    name: nginx-6cc9f5fc9f
    uid: 09ccf31f-4b05-4327-aed2-a64f4fed6293
  resourceVersion: "221300881"
  uid: 544cd5bc-a126-40d8-9b46-a9ff0ee1df55
spec:
  containers:
  - image: nginx:1.21.0
    imagePullPolicy: IfNotPresent
    name: nginx
  ...

Thought:I will now analyze the information obtained from the Pod and check for any possible security issues.

Action:
{
  "action": "Human",
  "action_input": "What are the main security concerns related to Kubernetes and cloud native technologies that I should be checking for?"
}

What are the main security concerns related to Kubernetes and cloud native technologies that I should be checking for?
image cve

Observation: image cve
Thought:I need to check if the nginx image used in this Pod has any known CVE vulnerabilities.

Action:
{
  "action": "Terminal",
  "action_input": "trivy image --severity HIGH,CRITICAL nginx:1.21.0"
}

Observation: 2023-03-25T11:23:13.756Z	INFO	Vulnerability scanning is enabled
2023-03-25T11:23:13.756Z	INFO	Secret scanning is enabled
2023-03-25T11:23:13.756Z	INFO	If your scanning is slow, please try '--scanners vuln' to disable secret scanning
2023-03-25T11:23:13.756Z	INFO	Please see also https://aquasecurity.github.io/trivy/v0.38/docs/secret/scanning/#recommendation for faster secret detection
2023-03-25T11:23:13.766Z	INFO	Detected OS: debian
2023-03-25T11:23:13.766Z	INFO	Detecting Debian vulnerabilities...
2023-03-25T11:23:13.787Z	INFO	Number of language-specific files: 0

nginx:1.21.0 (debian 10.10)
===========================
Total: 125 (HIGH: 95, CRITICAL: 30)
...
```

### Diagnose Problems for Pod

`kube-copilot diagnose POD [NAMESPACE]` will diagnose problems for a Pod.

Here is an example of diagnosing a problem Pod:

```sh
# kube-copilot diagnose nginx-944b5f9bd-ch67l

> Entering new AgentExecutor chain...
Question: What is the issue with the Pod nginx-944b5f9bd-ch67l in namespace default?
Thought: I need to gather more information about the Pod before attempting to diagnose any issues.
Action:
{
  "action": "Terminal",
  "action_input": "kubectl describe pod nginx-944b5f9bd-ch67l -n default"
}

Observation: Name:             nginx-944b5f9bd-ch67l
Namespace:        default
Start Time:       Sat, 25 Mar 2023 11:27:15 +0000
Labels:           app=nginx
                  pod-template-hash=944b5f9bd
Annotations:      <none>
Status:           Pending
Containers:
  nginx:
    Container ID:
    Image:          nginx:1.21.100
  ...

Thought:The Pod nginx-944b5f9bd-ch67l in namespace default is in Pending state, and the container inside it is in Waiting state with Reason as ImagePullBackOff. This indicates that the container is unable to pull the required image from the registry. Looking at the Event logs, we see that the image "nginx:1.21.100" could not be pulled due to a "NotFound" error. To resolve this issue, we need to check if the required image is available in the container registry and if it has the correct image name and tag.

Action:
{
  "action": "Terminal",
  "action_input": "kubectl get pods nginx-944b5f9bd-ch67l -n default -o yaml | grep image:"
}

Observation:   - image: nginx:1.21.100
  - image: nginx:1.21.100

Thought:The image name and tag in the Pod configuration file match the name and tag of the image that needs to be pulled. This means that the issue is not with the image configuration in the Pod file.

Action:
{
  "action": "Terminal",
  "action_input": "kubectl describe pod nginx-944b5f9bd-ch67l -n default | grep -i events -A 5"
}

Observation: Events:
  Type     Reason     Age                    From               Message
  ----     ------     ----                   ----               -------
  Normal   Pulling    5m6s (x4 over 6m43s)   kubelet            Pulling image "nginx:1.21.100"
  Warning  Failed     5m5s (x4 over 6m42s)   kubelet            Failed to pull image "nginx:1.21.100": rpc error: code = NotFound desc = failed to pull and unpack image "docker.io/library/nginx:1.21.100": failed to resolve reference "docker.io/library/nginx:1.21.100": docker.io/library/nginx:1.21.100: not found

Thought:From the Event logs, we can see that the image could not be found in the Docker registry. This indicates that the image has not been pushed to the required registry, or the image name and tag is incorrect. To resolve this issue, we need to either push the required image to the Docker registry or update the Pod configuration file with the correct image name and tag.

Action:
{
  "action": "Human",
  "action_input": "Can you check if the image nginx:1.21.100 is available in Dockerhub?"
}

Can you check if the image nginx:1.21.100 is available in Dockerhub?
No

Observation: No
Thought:The required image is not available in Dockerhub. We need to update the Pod configuration file with the correct image repository address and/or update the image tag to use an existing image in the repository. This should help the container to pull the required image and run the Pod without issues.

Action:
{
  "action": "Human",
  "action_input": "Do you know the URL of a container registry that contains the required image?"
}

Do you know the URL of a container registry that contains the required image?
No

Observation: No
Thought:We do not have any available container registry URL that contains the required image. We need to look for another image of the container with the correct image name and tag available on a container registry or build a new Docker image and push it to the registry with the correct name and tag.

Final Answer: The issue with the Pod nginx-944b5f9bd-ch67l in namespace default is that the required image "nginx:1.21.100" could not be found in the Docker registry. It is recommended to either push the required image to the Docker registry or update the Pod configuration file with the correct image repository address and/or update the image tag to use an existing image in the repository. If the issue still persists, look for another image of the container with the correct image name and tag available on a container registry or build a new Docker image and push it to the registry with the correct name and tag.

> Finished chain.
The issue with the Pod nginx-944b5f9bd-ch67l in namespace default is that the required image "nginx:1.21.100" could not be found in the Docker registry. It is recommended to either push the required image to the Docker registry or update the Pod configuration file with the correct image repository address and/or update the image tag to use an existing image in the repository. If the issue still persists, look for another image of the container with the correct image name and tag available on a container registry or build a new Docker image and push it to the registry with the correct name and tag.
```

### Execute Operations Based on Prompt Instructions

`kube-copilot execute INSTRUCTIONS` will execute operations based on prompt instructions.

Here is an example of querying the Pod consumed most CPU:

```sh
# kube-copilot execute "list the the Pod consumed most CPU with its current CPU consumption"

> Entering new AgentExecutor chain...
Thought: To complete the task, we need to use `Terminal` tool and run some kubectl commands to get the desired output.

Action:
{
  "action": "Terminal",
  "action_input": "kubectl top pods --all-namespaces | sort -k 3 -n -r | head -n 2 | tail -n 1"
}


Observation: pl                                  pl-etcd-0                                                        46m          257Mi

Thought:The output shows the Pod "pl" with its current CPU consumption at 257Mi. We need to further filter the output to show only the Pod name and its current CPU consumption.

Action:
{
  "action": "Terminal",
  "action_input": "kubectl top pods --all-namespaces | sort -k 3 -n -r | awk '{if (NR!=1){print $2,$3}}' | head -n 1"
}

Observation: pl-etcd-0 46m

Thought:The current output shows the Pod name and CPU consumption in millicores. We need to convert the CPU consumption to a more readable format.

Action:
{
  "action": "Terminal",
  "action_input": "kubectl top pods --all-namespaces | sort -k 3 -n -r | awk '{if (NR!=1){print $2,$3\"m\"}}' | awk '{printf \"%-50s %s\\n\" ,$1, $2/1000\"CPU\"}'"
}

Observation: pl-etcd-0                                          0.046CPU
pl-etcd-1                                          0.045CPU
pl-etcd-2                                          0.042CPU
...

Thought:The Pod that consumes most CPU with its current CPU consumption is "pl-etcd-0" with a consumption of 0.046CPU.

Final Answer: pl-etcd-0 with a consumption of 0.046CPU.

> Finished chain.
pl-etcd-0 with a consumption of 0.046CPU.
```

## Contribution

The project is opensource at github [feiskyer/kube-copilot](https://github.com/feiskyer/kube-copilot) with Apache License.

If you would like to contribute to the project, please follow these guidelines:

1. Fork the repository and clone it to your local machine.
2. Create a new branch for your changes.
3. Make your changes and commit them with a descriptive commit message.
4. Push your changes to your forked repository.
5. Open a pull request to the main repository.
