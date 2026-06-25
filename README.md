# Coffee Shop Microservices

Go microservices reference project focused on containerization, Kubernetes orchestration, CI/CD, and infrastructure-as-code practices.

## Architecture

| Service | Path | Protocol | Port | Purpose |
| --- | --- | --- | --- | --- |
| Products API | `products-api` | HTTP/JSON | `9090` | CRUD API for coffee products |
| Images API | `images-api` | HTTP | `9091` | Product image upload and download |
| Currency | `currency` | gRPC | `9092` | Currency conversion rate service |

Runtime flow:

```text
client -> products-api -> currency
client -> images-api
```

Products and images expose `/healthz`. Currency exposes the standard gRPC health service.

## Repository Layout

```text
.
├── .github/workflows/ci.yml   # GitHub Actions CI/CD pipeline
├── currency/                  # gRPC currency service
├── images-api/                # Product image service
├── k8s/                       # Kubernetes manifests and kustomization
├── products-api/              # Product REST API and generated SDK
├── docker-compose.yml         # Local multi-service runtime
├── go.work                    # Multi-module Go workspace
└── Makefile                   # Workspace commands
```

## Requirements

- Go `1.20`
- Docker with Buildx
- Docker Compose
- Kubernetes cluster for deployment tests, such as Docker Desktop Kubernetes
- `kubectl`

## Local Development

Run all tests from the repository root:

```bash
make test
```

`make test` runs tests across all Go workspace modules:

```bash
go test ./currency/... ./images-api/... ./products-api/...
```

The generated product SDK test expects the Products API to be reachable on `localhost:9090` when its cache is cold. The CI workflow starts required services before running tests.

## Docker

Each service has a multi-stage Dockerfile:

- build stage: `golang:1.20-alpine`
- runtime stage: `scratch`
- static Go binary
- non-root runtime user `65532:65532`

Build all local images:

```bash
docker compose build
```

Run locally with Compose:

```bash
docker compose up -d
```

Smoke test from the Compose network:

```bash
docker run --rm --network coffee-shop_default curlimages/curl:8.8.0 -fsS http://products-api:9090/products
```

Stop the stack:

```bash
docker compose down
```

## Configuration

| Variable | Default | Service | Description |
| --- | --- | --- | --- |
| `PRODUCTS_BIND_ADDR` | `127.0.0.1:9090` | products-api | HTTP bind address |
| `CURRENCY_ADDR` | `localhost:9092` | products-api | Currency gRPC endpoint |
| `IMAGES_BIND_ADDR` | `127.0.0.1:9091` | images-api | HTTP bind address |
| `IMAGE_STORE_PATH` | `./imagestore` | images-api | Local image storage path |
| `CURRENCY_BIND_ADDR` | `:9092` | currency | gRPC bind address |

The application currently has no database. `k8s/secrets.yaml` is intentionally empty until a real secret is required.

## Kubernetes

Kubernetes manifests live in `k8s/` and are managed by Kustomize:

```text
k8s/
├── namespace.yaml
├── configmaps.yaml
├── secrets.yaml
├── services.yaml
├── deployments.yaml
├── hpa.yaml
├── ingress.yaml
└── kustomization.yaml
```

Included platform features:

- dedicated `coffee-shop` namespace
- Deployments for all services
- ClusterIP Services for internal DNS
- ConfigMap-based runtime configuration
- HTTP and gRPC readiness/liveness probes
- CPU and memory requests/limits
- non-root container security contexts
- HorizontalPodAutoscalers
- Ingress routes for `/products` and `/images`

Validate manifests:

```bash
kubectl apply --dry-run=client -k k8s
```

Deploy:

```bash
kubectl apply -k k8s
```

Check rollout:

```bash
kubectl get pods -n coffee-shop
kubectl get svc -n coffee-shop
kubectl get hpa -n coffee-shop
```

Smoke test inside the cluster:

```bash
kubectl run curl \
  --namespace coffee-shop \
  --rm -it \
  --image=curlimages/curl:8.8.0 \
  --restart=Never -- \
  curl -fsS http://products-api-service:9090/products
```

## Container Registry

Kubernetes manifests reference GHCR images:

```text
ghcr.io/francisco3ferraz/coffee-shop/products-api:latest
ghcr.io/francisco3ferraz/coffee-shop/images-api:latest
ghcr.io/francisco3ferraz/coffee-shop/currency:latest
```

For a cloud cluster, push images through CI before applying manifests.

For Docker Desktop Kubernetes, local images may be used by temporarily setting image references back to `coffee-shop-*:local`, or by pulling the GHCR images after CI publishes them.

## CI/CD

GitHub Actions workflow: `.github/workflows/ci.yml`

On pushes to `main` or `master`, CI:

1. Checks out code.
2. Sets up Go `1.20`.
3. Starts local service dependencies.
4. Runs `make test`.
5. Logs in to GitHub Container Registry.
6. Builds and pushes all three optimized images.

Published image tags:

- `latest`
- commit SHA

## API Quick Reference

Products:

```bash
curl -fsS http://localhost:9090/products
curl -fsS http://localhost:9090/products/1
curl -fsS http://localhost:9090/healthz
```

Images:

```bash
curl -fsS http://localhost:9091/healthz
curl -fsS -X POST --data-binary @image.png http://localhost:9091/images/1/image.png
curl -fsS http://localhost:9091/images/1/image.png
```

## Operational Notes

- Currency rates are fetched from the European Central Bank at service startup.
- Images API uses local filesystem storage. Kubernetes uses `emptyDir`, so uploaded files are ephemeral.
- Ingress assumes an ingress controller with class `nginx`.
- HPA requires metrics-server in the cluster.
- The project is structured as multiple Go modules under one `go.work` workspace.
