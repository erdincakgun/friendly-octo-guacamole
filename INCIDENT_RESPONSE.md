1. If the **pending message count** is zero and the `Order Service` reports success, but the `Restaurant Service` has no record of the order; **technically**, where and how could these messages be disappearing? Provide two possible technical theories for this behavior.

- Messages were never written to the queue. The ordering service doesn't check whether the message was written to the queue. If producers were writing messages to the queue, we'd expect to see pending messages in a high traffic system, but we see zero. A TCP connection is broken. Messages are held in the buffer for a short time and then dropped by the OS. There are no application level checks like Heartbeat.

- Restaurant Service consumes messages from the queue and immediately sends ack before successfully processing them. During async processing, the service encounters external dependency failures such as expired SSL certificates, authentication tokens, or database credentials. These failures are logged as warnings rather than errors.

2. Why did our standard monitoring (HTTP status, CPU, Pending Message Count) fail to detect this incident? What was missing?

- Our monitoring stack not including business metrics like correlation between payment completed and restaurant confirmed. We are not collecting restaurant tablet application's metrics and traces. There is no external dependency health checks.

3. You have aligned with the business stakeholders, and the decision is made: **We must stop accepting new orders immediately** to prevent further financial discrepancies. However, the business goal is to keep the restaurant menus **fully browsable, searchable, and viewable** (Read-Only mode) so users don't abandon the app entirely. **What specific questions** do you ask the technical stakeholders to uncover the quickest and safest methods to achieve this targeted shutdown?

- Which endpoint or endpoints can be called to initiate a payment transaction? -> order service team
- Is there a maintenance page ready to use? -> mobile app team

4. The incident is resolved. You are now facilitating the **Postmortem** meeting. We don't want to rely on Customer Support to tell us the site is broken next time. What specific **Action Items** regarding **Alerting** or **SLIs** would you assign to the engineering teams to ensure we detect this specific scenario instantly in the future?

SLIs:

- Order Success Rate: Track correlation between "payment charged" events and "restaurant received order" events. Alert when ratio drops below 99% over 5 minutes.
- Consumer Processing Success Rate: Measure messages consumed vs messages successfully processed (written to DB). These numbers must match.

Alerts:

- DLQ depth > 0 for 2 minutes → P1 page
- Discrepancy between order service order count and restaurant service order count > 10 in 15 minutes → P1 page

Action Items:

| Action                                                                    | Owner                   | Priority |
| ------------------------------------------------------------------------- | ----------------------- | -------- |
| Implement order success rate SLI with tracing                             | Order Service Team      | P0       |
| Add DLQ monitoring and alerting                                           | Platform Team           | P0       |
| Create reconciliation job between Order and Restaurant DBs                | Data Engineering        | P1       |
| Instrument consumer to emit processing success/failure metrics            | Restaurant Service Team | P1       |
| Collect metrics and traces from restaurant tablet application             | Mobile Team             | P2       |
| Add external dependency health checks (SSL certs, tokens, DB connections) | Platform Team           | P2       |
