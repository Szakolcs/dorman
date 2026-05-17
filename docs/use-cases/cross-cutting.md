# Cross-Cutting Use Cases

Canonical requirements: [platform.md](../requirements/platform.md) (`FR-CC-*`). Integration contracts: [platform specification §6](../specifications/platform.md#6-cross-module-interaction-specification).

## UC-CC-01 Enforce Role-Based Access

- **Primary actor:** System
- **Goal:** Restrict actions by user role
- **Main flow:**
  1. User triggers action.
  2. System resolves user role and permission.
  3. System allows or denies operation.
- **Postconditions:** Decision is logged for sensitive actions.

## UC-CC-02 Deliver Notifications

- **Primary actor:** System
- **Goal:** Notify users about operationally important events
- **Main flow:**
  1. Event occurs (package, event update, moderation action).
  2. System checks notification preferences.
  3. System sends in-app notification.
- **Postconditions:** Delivery outcome is stored for audit.

## UC-CC-03 Observe Service Health

- **Primary actor:** Platform Operator
- **Goal:** Monitor health of app and dependencies
- **Main flow:**
  1. Operator queries health endpoint.
  2. System returns status for API and DB dependencies.
  3. Operator investigates if component is degraded.
- **Postconditions:** Operational awareness is maintained.
