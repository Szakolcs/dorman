# Cross-Cutting User Stories

Canonical requirements: [platform.md](../requirements/platform.md) (`FR-CC-*`). Integration contracts: [platform specification §6](../specifications/platform.md#6-cross-module-interaction-specification).

## CC-001 Role-Based Access Control

- **As a** system administrator
- **I want** role-based permissions across modules
- **So that** users only access features they are allowed to use

### Acceptance Criteria

- Roles include admin, office worker, doorman, and tenant.
- Unauthorized actions return a clear permission error.
- Permission checks are audited for sensitive actions.

## CC-002 Audit Trail

- **As a** compliance-focused administrator
- **I want** critical actions to be logged with actor and timestamp
- **So that** operational decisions are traceable

### Acceptance Criteria

- Logs are stored for room assignments, guest access, and moderation actions.
- Audit records are searchable by user, action type, and date range.
- Audit records are immutable for non-admin roles.

## CC-003 Notification Preferences

- **As a** tenant
- **I want** to manage notification preferences
- **So that** I receive relevant updates without spam

### Acceptance Criteria

- Preferences support package alerts, event updates, and chat notifications.
- Opt-out is available for non-critical announcements.
- Critical dorm safety notices cannot be disabled.

## CC-004 Service Health Visibility

- **As a** platform operator
- **I want** basic service health endpoints and checks
- **So that** outages and degraded services are detected early

### Acceptance Criteria

- Health endpoint reports service and database status.
- Failing dependencies return actionable diagnostics.
- Health checks are documented for deployment monitoring.
