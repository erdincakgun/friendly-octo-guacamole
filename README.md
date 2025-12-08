# Centralized Observability Platform - SRE Challenge

A proof-of-concept implementation of a centralized observability platform for a food delivery platform, featuring unified telemetry collection using OpenTelemetry, resilient data pipelines with Kafka, and a scalable storage backend.

## Tech Stack

| Component              | Technology                  | Chart/Version                                                                  |
| ---------------------- | --------------------------- | ------------------------------------------------------------------------------ |
| **Application**        | Go (mock-service)           | friendly-octo-guacamole/1.3.0                                                  |
| **Ingress**            | NGINX Ingress Controller    | ingress-nginx/4.14.1                                                           |
| **Log Collection**     | Fluent Bit                  | fluent/fluent-bit/0.54.0 (image: 4.1.1)                                        |
| **Telemetry Pipeline** | OpenTelemetry Collector     | open-telemetry/opentelemetry-collector/0.140.1                                 |
| **Message Broker**     | Apache Kafka (Strimzi)      | strimzi/strimzi-kafka-operator/0.49.1                                          |
| **Kafka UI**           | Kafbat UI                   | kafka-ui/kafka-ui/0.7.3                                                        |
| **Logs Storage**       | Grafana Loki (Distributed)  | grafana/loki/6.46.0                                                            |
| **Traces Storage**     | Grafana Tempo (Distributed) | grafana/tempo-distributed/1.57.0                                               |
| **Metrics**            | Prometheus + Thanos         | prometheus-community/kube-prometheus-stack/79.12.0, stevehipwell/thanos/1.22.0 |
| **Object Storage**     | MinIO                       | minio/operator/7.1.1, minio/tenant/7.1.1                                       |
| **Visualization**      | Grafana                     | grafana/grafana/10.3.0                                                         |
| **Metrics Server**     | Kubernetes Metrics Server   | metrics-server/metrics-server/3.13.0                                           |

### Prerequisites

- Kubernetes cluster (Docker Desktop, Minikube, Kind, or K3d)
- kubectl
- Helm 3.x
- Helmfile
- SOPS + age key (for secrets decryption)

### DNS

```
127.0.0.1 friendly-octo-guacamole.com
127.0.0.1 grafana.friendly-octo-guacamole.com
127.0.0.1 kafka-ui.friendly-octo-guacamole.com
127.0.0.1 minio-console.friendly-octo-guacamole.com
```

### Deploy the Stack

```bash
# Preview changes
helmfile --environment production diff

# Apply changes
helmfile --environment production apply --skip-deps

# Deploy everything
helmfile --environment production sync
```

### Access the Services

| Service       | URL                                              |
| ------------- | ------------------------------------------------ |
| Menu API      | http://friendly-octo-guacamole.com               |
| Grafana       | http://grafana.friendly-octo-guacamole.com       |
| Kafka UI      | http://kafka-ui.friendly-octo-guacamole.com      |
| MinIO Console | http://minio-console.friendly-octo-guacamole.com |

## Application Endpoints

The `mock-service` simulates a Menu API with intentional failure scenarios:

| Endpoint         | Method | Description     | Response Codes            |
| ---------------- | ------ | --------------- | ------------------------- |
| `/health`        | GET    | Health check    | `200` (always)            |
| `/api/menu`      | GET    | List menu items | `200`, `500` (10% chance) |
| `/api/menu/{id}` | GET    | Get menu item   | `200`, `404`              |

### Testing Endpoints

```bash
# Health check
curl -i http://friendly-octo-guacamole.com/health

# List menu items (may randomly return 500)
curl -i http://friendly-octo-guacamole.com/api/menu

# Get specific menu item
curl -i http://friendly-octo-guacamole.com/api/menu/1

# Non-existent item (returns 404)
curl -i http://friendly-octo-guacamole.com/api/menu/999
```

## System Design Decisions

### 1. Unified & Standardized Collection (OTLP)

- **OpenTelemetry Collector** handles all telemetry data using the OTLP protocol
- Avoids vendor lock-in by using standard protocols
- Single collector deployment handles logs, metrics, and traces

### 2. Resilience via Kafka

- **Kafka (Strimzi)** acts as a buffer between collection and storage layers
- If `central-observability` backends are temporarily unavailable, data is retained in Kafka
- OpenTelemetry Collector consumes from Kafka with `initial_offset: earliest` ensuring no data loss

### 3. Highly Available & Scalable Central Stack

- **Loki**: Distributed mode with separate read/write paths
- **Tempo**: Distributed mode for trace storage
- **Thanos**: HA metrics storage with long-term retention
- **MinIO**: S3-compatible object storage for cost-effective 1-year retention

## Structured Logging

All logs are output in JSON format:

```json
{
  "timestamp": "2025-11-22T10:30:45Z",
  "level": "INFO",
  "method": "GET",
  "path": "/api/menu",
  "status_code": 200,
  "duration_ms": 2.34,
  "request_id": "req-1234567890-1",
  "message": "Listed 5 menu items"
}
```

### Log Fields

| Field          | Description                             |
| -------------- | --------------------------------------- |
| `timestamp`    | RFC3339 formatted timestamp             |
| `level`        | Log level (INFO, WARN, ERROR)           |
| `method`       | HTTP method                             |
| `path`         | Request path                            |
| `status_code`  | HTTP response code                      |
| `duration_ms`  | Request processing time in milliseconds |
| `request_id`   | Unique request identifier               |
| `menu_item_id` | Menu item ID (when applicable)          |
| `message`      | Human-readable message                  |
| `error`        | Error details (only on failures)        |

## Grafana Dashboard

The unified dashboard (`Menu API - Unified Dashboard`) provides insights for multiple stakeholders:

### Customer Experience View

- Service Availability percentage
- Menu Request Success Rate

### Business View

- Request volume over time
- Error rates by endpoint

### Technical View

- Response time distribution (Avg/Max)
- Requests by endpoint breakdown
- Distributed tracing visualization

Access the dashboard at: http://grafana.friendly-octo-guacamole.com/d/menu-api-unified/

## Project Structure

```
.
├── charts/
│   └── friendly-octo-guacamole/    # Helm chart for the application
├── manifests/
│   └── production/
│       └── kafka.yaml              # Kafka cluster manifest (Strimzi)
├── values/
│   └── production/
│       ├── fluent-bit.yaml         # Log collection config
│       ├── friendly-octo-guacamole.yaml
│       ├── grafana.yaml            # Grafana + dashboard JSON
│       ├── kube-prometheus-stack.yaml
│       ├── loki.yaml               # Distributed Loki config
│       ├── minio.yaml              # Object storage config
│       ├── open-telemetry-collector.yaml
│       ├── tempo.yaml              # Distributed Tempo config
│       └── thanos.yaml             # HA metrics config
├── helmfile.yaml                   # Declarative Helm releases
├── main.go                         # Application source
└── CHALLENGE.md                    # SRE Challenge description
```

## Task 1: Incident Response Answers

See [INCIDENT_RESPONSE.md](./INCIDENT_RESPONSE.md) for detailed answers to the incident scenario questions.
