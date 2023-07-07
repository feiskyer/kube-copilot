# Helm Charts for kube-copilot

## Install with OpenAI

```sh
helm install kube-copilot kube-copilot \
  --repo https://feisky.xyz/kube-copilot \
  --set openai.apiModel=gpt-4 \
  --set openai.apiKey=$OPENAI_API_KEY
```

## Install with Azure OpenAI Service

```sh
helm install kube-copilot kube-copilot \
  --repo https://feisky.xyz/kube-copilot \
  --set openai.apiModel=gpt-4 \
  --set openai.apiKey=$OPENAI_API_KEY \
  --set openai.apiBase=$OPENAI_API_BASE
```

## Enable Google Search

```sh
helm install kube-copilot kube-copilot \
  --repo https://feisky.xyz/kube-copilot \
  --set openai.apiModel=gpt-4 \
  --set openai.apiBase=$OPENAI_API_BASE \
  --set openai.apiKey=$OPENAI_API_KEY \
  --set google.apiKey=$GOOGLE_API_KEY \
  --set google.cseId=$GOOGLE_CSE_ID
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| google.apiKey | string | `""` | Required when using Google Search |
| google.cseId | string | `""` | Required when using Google Search |
| openai.apiBase | string | `""` | Required when using Azure OpenAI Service |
| openai.apiKey | string | `""` | Required |
| openai.apiModel | string | `"gpt-4"` |  |
| image.repository | string | `"ghcr.io/feiskyer/kube-copilot"` |  |
| image.tag | string | `"latest"` |  |
| imagePullSecrets | list | `[]` |  |
| ingress.annotations | object | `{}` |  |
| ingress.className | string | `""` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts[0].host | string | `"chart-example.local"` |  |
| ingress.hosts[0].paths[0].path | string | `"/"` |  |
| ingress.hosts[0].paths[0].pathType | string | `"ImplementationSpecific"` |  |
| ingress.tls | list | `[]` |  |
| nodeSelector | object | `{}` |  |
| replicaCount | int | `1` |  |
| resources | object | `{}` |  |
| service.port | int | `80` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `"kube-copilot"` |  |
| tolerations | list | `[]` |  |
