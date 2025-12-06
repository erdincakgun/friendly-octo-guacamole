# SRE Challenge

## Task 1

#### Context & Architecture

You are a SRE at a high-traffic food delivery platform. You are the Incident Commander on call. 

**Simplified Request Flow:** `Mobile App` → `API Gateway` → `Order Service` → `Message Broker` → `Restaurant Service`

- **Order Service:** Charges the user and publishes the order event to the broker.
- **Restaurant Service:** Consumes events from the broker and pushes them to the restaurant's tablet.

#### The Incident

**Time:** Tuesday, 21:00 (Regular traffic). 
**Constraint:** There have been **no recent deployments**, configuration changes, or infrastructure updates in the last 24 hours.

**Symptoms:** The Customer Support Lead pages you urgently:

> _"We are facing a critical issue. Customers are being charged, the app shows 'Order Received', but restaurants are not receiving the orders. We are effectively taking money for food we aren't delivering."_

**Your Dashboard Status:**
- All services report **Healthy**.
- Error rates are negligible (< 0.1%).
- System resources (CPU/RAM) are stable.
- The Message Broker shows **zero pending messages** (no backlog).

---
Please answer the following questions. We are interested in your **reasoning process**, **prioritization**, and **communication style**, rather than specific tool commands.

1. If the **pending message count** is zero and the `Order Service` reports success, but the `Restaurant Service` has no record of the order; **technically**, where and how could these messages be disappearing? Provide two possible technical theories for this behavior.

2. Why did our standard monitoring (HTTP status, CPU, Pending Message Count) fail to detect this incident? What was missing?

3. You have aligned with the business stakeholders, and the decision is made: **We must stop accepting new orders immediately** to prevent further financial discrepancies. However, the business goal is to keep the restaurant menus **fully browsable, searchable, and viewable** (Read-Only mode) so users don't abandon the app entirely. **What specific questions** do you ask the technical stakeholders to uncover the quickest and safest methods to achieve this targeted shutdown?

4. The incident is resolved. You are now facilitating the **Postmortem** meeting. We don't want to rely on Customer Support to tell us the site is broken next time. What specific **Action Items** regarding **Alerting** or **SLIs** would you assign to the engineering teams to ensure we detect this specific scenario instantly in the future?

___

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
4. **(Bonus)** The application currently lacks **Distributed Tracing**. It would be a significant "nice-to-have" if you could enable tracing for this service. We leave the implementation details  entirely up to your judgment. We just want to be able to visualize a trace for an HTTP request in the observability platform.