## Task 2

#### Context

In our real-world scenario, we run multiple Kubernetes clusters. We need a **Centralized Observability Platform** that aggregates data from all workloads.

To test your architectural and implementation skills, we have provided a specific **Go Application (`mock-service`)** that simulates a "Menu API" for our food delivery platform. This application is known to have intentional failure scenarios and emits **Structured JSON Logs**.

#### The Challenge

Design and implement a proof-of-concept (PoC) using a **Single Kubernetes Cluster** (Local Minikube/Kind/K3d or Cloud). _Note: We do not expect a CI/CD pipeline. Manual deployments (kubectl/helm install) are perfectly acceptable for this task._

**Required Namespace Structure:**

1. **`production`**: Deploy the provided `mock-service` application here.

2. **`monitoring` (The Collection Layer):** Deploy your collection stack here. This represents the monitoring setup running inside the workload cluster.

3. **`central-observability` (The Hub Layer):** Deploy your centralized storage and visualization stack here.

#### System Requirements

While you have the freedom to choose the specific tools, your solution must satisfy the following architectural goals:

1. **Unified & Standardized Collection:** We want to avoid vendor lock-in. The collection layer should ideally leverage **standard protocols (e.g., OTLP)** to handle Metrics, Logs, and Traces in a unified manner.

2. **Resilience:** The collection layer should be resilient to network failures. If the connection to `central-observability` is lost temporarily, we should **not lose data**.

3. **Central Stack Architecture:** The central backend must be **Highly Available** and **Scalable** to handle growing data volumes. It should support **1-year retention** for compliance, but the storage strategy must be **cost-effective**.

---

## Deliverables

We expect a solution that is **fully reproducible**. During the case review session, we should be able to see this setup working live.

Please provide a link to a **Git Repository** containing:

1. A Markdown file containing your detailed answers to the incident scenario questions.

2. Provide Infrastructure as Code files (Complete Helm Charts, Kustomize files, or Terraform manifests, etc.).

   - **Note:** Anyone cloning this repo should be able to spin up the entire stack (App + Observability) with a few standard commands (e.g., `kubectl apply -f ...`, `helm install ...`, or a `make up` script). Avoid manual steps or "magic" local configurations.

3. The application does not natively expose business metrics, yet we require them. You need to **derive and visualize these insights** using your observability stack. We expect a unified Grafana dashboard that includes these generated business metrics. The dashboard must be designed so that **Customer Experience, Business, and Technical teams** can use it to understand the issue from their respective perspectives. The Grafana Dashboard JSON file must be committed to the repository.

4. **(Bonus)** The application currently lacks **Distributed Tracing**. It would be a significant "nice-to-have" if you could enable tracing for this service. We leave the implementation details entirely up to your judgment. We just want to be able to visualize a trace for an HTTP request in the observability platform.
