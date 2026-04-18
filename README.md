# Kubernetes Admission Webhook

A Kubernetes admission webhook server written in Go that validates and mutates pod creation requests.

## What it does

**Validating webhook** — rejects pods that:
- Use the `:latest` image tag
- Run as root (`runAsUser: 0`)

**Mutating webhook** — adds the label `admitted-by: webhook` to every pod.

## Project structure

```
.
├── main.go                        # Webhook server (validate + mutate handlers)
├── types.go                       # AdmissionReview types
├── Dockerfile                     # Multi-stage build
├── go.mod
├── certs/
│   └── gen-certs.sh               # Script to generate TLS certs with SAN
└── k8s/
    ├── deployment.yaml
    ├── service.yaml
    ├── validating-webhook.yaml
    └── mutating-webhook.yaml
```

## Prerequisites

- Kubernetes cluster (e.g. Minikube)
- `kubectl` configured
- `openssl`
- Docker

## Setup

### 1. Generate TLS certificates

```bash
cd certs

# Generate CA
openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 365 -key ca.key -subj "/CN=webhook-ca" -out ca.crt

# Generate server cert with SAN
SERVICE=webhook-service NAMESPACE=default bash gen-certs.sh
```

### 2. Create the TLS secret

```bash
kubectl create secret tls webhook-tls \
  --cert=certs/tls.crt \
  --key=certs/tls.key \
  -n default
```

### 3. Set the caBundle in webhook configs

```bash
CA_BUNDLE=$(base64 < certs/ca.crt | tr -d '\n')

kubectl patch validatingwebhookconfiguration webhook-validating \
  --type='json' \
  -p="[{\"op\":\"replace\",\"path\":\"/webhooks/0/clientConfig/caBundle\",\"value\":\"${CA_BUNDLE}\"}]"

kubectl patch mutatingwebhookconfiguration webhook-mutating \
  --type='json' \
  -p="[{\"op\":\"replace\",\"path\":\"/webhooks/0/clientConfig/caBundle\",\"value\":\"${CA_BUNDLE}\"}]"
```

### 4. Build and push the Docker image

```bash
docker build -t <your-image>:v1.0.0 .
docker push <your-image>:v1.0.0
```

Update the image in `k8s/deployment.yaml` then apply all manifests:

```bash
kubectl apply -f k8s/
```

## Testing

```bash
# Should be allowed (pinned tag)
kubectl run test-ok --image=nginx:1.21

# Should be rejected (latest tag)
kubectl run test-bad --image=nginx:latest

# Should be rejected (runs as root)
kubectl run test-root --image=nginx:1.21 --overrides='{"spec":{"containers":[{"name":"test-root","image":"nginx:1.21","securityContext":{"runAsUser":0}}]}}'
```

## Notes

- The webhook server listens on port `8443` (TLS)
- The Kubernetes service exposes it on port `443`
- TLS certs must include a SAN for `webhook-service.default.svc`
- Never commit `*.key` files — they are excluded via `certs/.gitignore`
