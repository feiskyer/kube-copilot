# Run in Kubernetes

## Web UI with Helm (recommended)

### Option 1: Install with OpenAI

```sh
export OPENAI_API_KEY="<replace-this>"

helm install kube-copilot kube-copilot \
  --repo https://feisky.xyz/kube-copilot \
  --set openai.apiModel=gpt-4 \
  --set openai.apiKey=$OPENAI_API_KEY
```

### Option 2: Install with Azure OpenAI Service

```sh
export OPENAI_API_KEY="<replace-this>"
export OPENAI_API_BASE="<replace-this>"

helm install kube-copilot kube-copilot \
  --repo https://feisky.xyz/kube-copilot \
  --set openai.apiModel=gpt-4 \
  --set openai.apiKey=$OPENAI_API_KEY \
  --set openai.apiBase=$OPENAI_API_BASE
```

### Option 3: Azure OpenAI Service + Google Search

```sh
export OPENAI_API_KEY="<replace-this>"
export OPENAI_API_BASE="<replace-this>"
export GOOGLE_API_KEY="<replace-this>"
export GOOGLE_CSE_ID="<replace-this>"

helm install kube-copilot kube-copilot \
  --repo https://feisky.xyz/kube-copilot \
  --set openai.apiModel=gpt-4 \
  --set openai.apiBase=$OPENAI_API_BASE \
  --set openai.apiKey=$OPENAI_API_KEY \
  --set google.apiKey=$GOOGLE_API_KEY \
  --set google.cseId=$GOOGLE_CSE_ID
```

## Manually Install for CLI mode

Create RBAC rule and binding:

```sh
kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kube-copilot-reader
rules:
- apiGroups:
  - '*'
  resources:
  - '*'
  verbs:
  - 'get'
  - 'list'
- nonResourceURLs:
  - '*'
  verbs:
  - 'get'
  - 'list'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kube-copilot
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kube-copilot-reader
subjects:
- kind: ServiceAccount
  name: kube-copilot
  namespace: default
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kube-copilot
  namespace: default
automountServiceAccountToken: true
EOF
```

Create secret:

```sh
kubectl create secret generic kube-copilot-auth \
    --from-literal=OPENAI_API_TYPE=${OPENAI_API_TYPE} \
    --from-literal=OPENAI_API_KEY=${OPENAI_API_KEY} \
    --from-literal=OPENAI_API_BASE=${OPENAI_API_BASE}
```

Run:

```sh
kubectl run -it --rm copilot \
  --restart=Never \
  --image=ghcr.io/feiskyer/kube-copilot \
  --overrides='
{
  "spec": {
    "serviceAccountName": "kube-copilot",
    "containers": [
      {
        "name": "copilot",
        "image": "ghcr.io/feiskyer/kube-copilot",
        "env": [
          {
            "name": "OPENAI_API_KEY",
            "valueFrom": {
              "secretKeyRef": {
                "name": "kube-copilot-auth",
                "key": "OPENAI_API_KEY"
              }
            }
          },
          {
            "name": "OPENAI_API_BASE",
            "valueFrom": {
              "secretKeyRef": {
                "name": "kube-copilot-auth",
                "key": "OPENAI_API_BASE"
              }
            }
          },
          {
            "name": "OPENAI_API_TYPE",
            "valueFrom": {
              "secretKeyRef": {
                "name": "kube-copilot-auth",
                "key": "OPENAI_API_TYPE"
              }
            }
          }
        ]
      }
    ]
  }
}' \
  -- execute --verbose 'What Pods are using max memory in the cluster'
```

## Manually install with default service account

If the default service account is already configured with the expected RBAC bindings, you can use this simpler method to run:

```sh
kubectl run -it --rm copilot \
  --env="OPENAI_API_KEY=$OPENAI_API_KEY" \
  --restart=Never \
  --image=ghcr.io/feiskyer/kube-copilot \
  -- execute --verbose 'What Pods are using max memory in the cluster'
```
