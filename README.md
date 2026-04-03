# EventHub — K8s Multi-Tenant Learning Platform

A complete microservices platform for learning Kubernetes workload management,
Helm chart authoring, Dockerfile patterns, and platform engineering.

## Architecture

```
platform-services/          ← Central PostgreSQL (users, orgs, roles)
  ├── postgresql              StatefulSet + PVC
  ├── pgbouncer               Connection pooler
  └── tenant-onboarding       Automated provisioning

tenant-{name}/              ← Per-tenant namespace
  ├── event-service           Go + MySQL (REST API)
  ├── live-dashboard          Node.js + MongoDB (WebSocket dashboard)
  └── analytics-worker        Python + Redis (Background jobs)
```

## Quick Start

### Prerequisites
- Kubernetes cluster (minikube, kind, or cloud)
- Helm 3.x
- kubectl
- Docker

### 1. Deploy Platform Services
```bash
kubectl create namespace platform-services
helm install central-pg ./platform/postgresql/helm -n platform-services
helm install pgbouncer ./platform/pgbouncer/helm -n platform-services
```

### 2. Onboard a Tenant
```bash
helm install tenant-alpha ./platform/tenant-onboarding/helm \
  -f values/tenant-alpha.yaml
```

### 3. Deploy Applications
```bash
make build-all
helm install event-service ./apps/event-service/helm \
  -n tenant-alpha -f values/tenant-alpha.yaml
helm install live-dashboard ./apps/live-dashboard/helm \
  -n tenant-alpha -f values/tenant-alpha.yaml
helm install analytics-worker ./apps/analytics-worker/helm \
  -n tenant-alpha -f values/tenant-alpha.yaml
```

## Learning Path
1. Central DB + tenant provisioning
2. Network policies for isolation
3. PgBouncer connection pooling
4. Resource quotas and limits
5. Schema migrations via Jobs
6. Monitoring with Prometheus/Grafana
7. Failure injection drills
8. Self-service onboarding automation
# platform-apps
