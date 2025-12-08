1. If the **pending message count** is zero and the `Order Service` reports success, but the `Restaurant Service` has no record of the order; **technically**, where and how could these messages be disappearing? Provide two possible technical theories for this behavior.

- Messages were never written to the queue. The ordering service doesn't check whether the message was written to the queue. If producers were writing messages to the queue, we'd expect to see pending messages in a high traffic system, but we see zero. A TCP connection is broken. Messages are held in the buffer for a short time and then dropped by the OS. There are no application level checks like Heartbeat.

- Restaurant Service consumes messages from the queue and immediately sends ack before successfully processing them. During async processing, the service encounters external dependency failures such as expired SSL certificates, authentication tokens, or database credentials. These failures are logged as warnings rather than errors.

2. Why did our standard monitoring (HTTP status, CPU, Pending Message Count) fail to detect this incident? What was missing?

- Our monitoring stack not including business metrics like correlation between payment completed and restaurant confirmed. We are not collecting restaurant tablet application's metrics and traces. There is no external dependency health checks.

3. You have aligned with the business stakeholders, and the decision is made: **We must stop accepting new orders immediately** to prevent further financial discrepancies. However, the business goal is to keep the restaurant menus **fully browsable, searchable, and viewable** (Read-Only mode) so users don't abandon the app entirely. **What specific questions** do you ask the technical stakeholders to uncover the quickest and safest methods to achieve this targeted shutdown?

- Which endpoint or endpoints can be called to initiate a payment transaction? -> order service team
- Is there a maintenance page ready to use? -> mobile app team

1. The incident is resolved. You are now facilitating the **Postmortem** meeting. We don't want to rely on Customer Support to tell us the site is broken next time. What specific **Action Items** regarding **Alerting** or **SLIs** would you assign to the engineering teams to ensure we detect this specific scenario instantly in the future?
