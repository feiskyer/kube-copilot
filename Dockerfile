# build stage
FROM golang:alpine AS builder
ADD . /go/src/github.com/feiskyer/kube-copilot
RUN cd /go/src/github.com/feiskyer/kube-copilot && \
    apk update && apk add --no-cache gcc musl-dev openssl && \
    CGO_ENABLED=0 go build -o _out/kube-copilot ./cmd/kube-copilot

# Final image
FROM alpine
# EXPOSE 80
WORKDIR /

RUN apk add --update curl wget python3 py3-pip curl && \
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && mv kubectl /usr/local/bin && \
    curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin v0.48.1 && \
    rm -rf /var/cache/apk/* && \
    mkdir -p /etc/kube-copilot

COPY --from=builder /go/src/github.com/feiskyer/kube-copilot/_out/kube-copilot /usr/local/bin/

USER copilot
ENTRYPOINT [ "/usr/local/bin/kube-copilot" ]
