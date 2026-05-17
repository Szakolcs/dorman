# Platform and Cross-Cutting Requirements

## Purpose

This document defines requirements for shared platform capabilities that span product modules: identity and role-based access control, audit logging, tenant notification preferences, and operational health visibility. It aligns with [platform-features.md](../features/platform-features.md), [cross-cutting use cases](../use-cases/cross-cutting.md), and [cross-cutting user stories](../user-stories/cross-cutting.md).

Module-specific requirements (for example `FR-AO-014`, `FR-DM-012`, `FR-FM-015`, `FR-CM-012`) restate enforcement expectations inside each bounded context; this document is the canonical source for cross-cutting behavior.

## Scope

The platform layer covers:

- Authenticated principals (staff and tenant) and role resolution
- Server-side authorization for privileged actions across modules
- Immutable audit records for sensitive operations
- In-app notification delivery and per-tenant preference controls
- Service and dependency health reporting for operators

The platform shares boundaries with:

- **Administration and Office**: staff principals, housing and publication entities referenced by audit targets and notification payloads
- **Doorman**: doorman role, package notification fan-out, guest and access audit events
- **Forum**: tenant principals, moderation audit, optional notification events from polls and events
- **Chat**: tenant principals, flat membership driven externally, optional message notification events

## Actors and Roles

- **Administrator**: manages roles, reviews audit history, configures non-critical notification defaults where policy allows
- **Office Worker / Doorman / Tenant**: operate within module UIs subject to role grants
- **Platform Operator**: monitors health endpoints and dependency status
- **System**: resolves permissions, writes audit rows, enqueues notifications, evaluates preference gates

## Functional Requirements

### Identity and Authorization

#### FR-CC-001 Role-Based Access Control

The system shall enforce role-based permissions for all privileged module actions.

Acceptance:

- roles include at minimum: administrator, office worker, doorman, and tenant (plus scoped grants such as Student Consult Organizer where product defines them)
- unauthorized actions return a clear permission error without partial side effects
- authorization decisions use server-resolved role data; client-supplied role hints are non-authoritative
- permission checks run at API and application service boundaries before domain mutations

#### FR-CC-002 Principal Context

The system shall bind each request to an authenticated principal with stable identifiers for audit and authorization.

Acceptance:

- staff and tenant sessions resolve to `User.id` and, for tenants, linked `Tenant.id` where applicable
- inactive or revoked users cannot perform mutations
- development-only staff stubs (if enabled) are clearly separated from production authentication paths

### Audit and Traceability

#### FR-CC-003 Sensitive Action Audit

The system shall record audit events for sensitive operations across modules.

Acceptance:

- minimum fields: actor user id (nullable only for automated system actions), action identifier, target type, target id, outcome, timestamp
- examples include room assignments, guest access denials, package status overrides, forum moderation, and privileged chat actions when implemented
- audit records are append-only for non-administrator roles
- audit history is searchable by actor, action, target type, and date range for authorized staff

#### FR-CC-004 Module Audit Delegation

Each product module shall emit platform audit events for its own sensitive mutations rather than duplicating audit stores.

Acceptance:

- domain tables may retain fine-grained history (for example `TicketStatusChange`, `ForumModerationAction`) in addition to `AuditEvent`
- module services call a shared audit writer with consistent action naming conventions (`<module>.<verb>`)

### Notifications

#### FR-CC-005 Notification Preferences

The system shall allow tenants to manage notification preferences for non-critical categories.

Acceptance:

- preferences cover at minimum: package alerts, event or activity updates, and chat message notifications
- tenants may opt out of non-critical categories
- critical dorm safety notices cannot be disabled
- preference reads are applied before enqueueing in-app notification delivery

#### FR-CC-006 Event-Driven In-App Notifications

The system shall deliver in-app notifications triggered by domain events from product modules.

Acceptance:

- modules emit logical events (for example package ready, poll closed, chat message created) without owning delivery infrastructure
- delivery respects tenant preferences and records outcome for troubleshooting
- notification payloads reference stable entity ids for deep links where product defines routes

### Operations

#### FR-CC-007 Service Health Visibility

The system shall expose health endpoints suitable for deployment monitoring.

Acceptance:

- health reports API process status and database connectivity
- degraded dependencies return actionable status codes or body fields for operators
- health check contract is documented for deployment runbooks

## Non-Functional Requirements

#### NFR-CC-001 Security

- audit and role tables are writable only through application services
- health endpoints do not leak secrets or connection strings

#### NFR-CC-002 Reliability

- audit write failures on sensitive mutations must fail the business transaction or be retried with explicit policy (default: fail closed for compliance-critical actions)

#### NFR-CC-003 Performance

- authorization checks should be O(1) per request via cached role resolution where safe
- audit search endpoints support pagination and indexed filters on `occurred_at`, `actor_user_id`, and `action`

## Inter-Module Integration Requirements

| Integration | Platform responsibility | Consumer module behavior |
|-------------|-------------------------|---------------------------|
| Staff auth | Resolve `User` + staff roles | Administration, doorman, forum staff actions |
| Tenant auth | Resolve `User` + `Tenant` | Forum, chat tenant actions |
| Audit | Persist `AuditEvent` | All modules emit on sensitive writes |
| Notifications | Preference gate + in-app delivery | Doorman package events, forum/chat optional hooks |
| Health | `/health` or equivalent | External probes only |

## Business Rules

- A principal may hold multiple roles; effective permissions are the union of grants unless explicitly restricted by policy.
- Audit records are never updated in place; corrections are new rows or domain-specific compensating entries.
- Notification preference defaults favor opt-in for operational categories until the tenant changes settings.

## Out of Scope

- Module-owned domain tables (packages, forum posts, chat messages)
- External push providers (email/SMS gateways) beyond hooks noted in module requirements
- Full SSO or federation (may be added later; MVP uses application-managed credentials)

## Traceability Matrix

| Requirement | Source alignment |
|---|---|
| FR-CC-001, FR-CC-002 | CC-001, UC-CC-01, Authentication and Role Permissions features |
| FR-CC-003, FR-CC-004 | CC-002, UC-CC-02 (audit aspects), Audit and Traceability features |
| FR-CC-005, FR-CC-006 | CC-003, UC-CC-02, Notification System features |
| FR-CC-007 | CC-004, UC-CC-03, Health and Monitoring features |
| Module FR-AO-014/015 | Delegates to FR-CC-001, FR-CC-003 |
| Module FR-DM-012/013 | Delegates to FR-CC-001, FR-CC-003 |
| Module FR-FM-015/016 | Delegates to FR-CC-001, FR-CC-003 |
| Module FR-CM-012/013 | Delegates to FR-CC-001, FR-CC-003 |
