`Mobile App` → `API Gateway` → `Order Service` → `Message Broker` → `Restaurant Service`

- **Order Service:** 
  - Charges the user
  - Publishes the order event

- **Restaurant Service:**
  - Consumes events
  - Pushes events to the restaurant's tablet.

- Regular traffic
- No changes in the last 24 hours
  
- Customer Support Lead pages..
- Customers are being charged
- the app shows 'Order Received'
- restaurants are not receiving the orders

- All services report **Healthy**.
- Error rates are negligible (< 0.1%).
- System resources (CPU/RAM) are stable.
- The Message Broker shows **zero pending messages** (no backlog).

---

- **reasoning process**
- **prioritization**
- **communication style**
- rather than specific tool commands

1. If the **pending message count** is zero and the `Order Service` reports success, but the `Restaurant Service` has no record of the order; **technically**, where and how could these messages be disappearing? Provide two possible technical theories for this behavior.

- 

1. Why did our standard monitoring (HTTP status, CPU, Pending Message Count) fail to detect this incident? What was missing?

2. You have aligned with the business stakeholders, and the decision is made: **We must stop accepting new orders immediately** to prevent further financial discrepancies. However, the business goal is to keep the restaurant menus **fully browsable, searchable, and viewable** (Read-Only mode) so users don't abandon the app entirely. **What specific questions** do you ask the technical stakeholders to uncover the quickest and safest methods to achieve this targeted shutdown?

3. The incident is resolved. You are now facilitating the **Postmortem** meeting. We don't want to rely on Customer Support to tell us the site is broken next time. What specific **Action Items** regarding **Alerting** or **SLIs** would you assign to the engineering teams to ensure we detect this specific scenario instantly in the future?