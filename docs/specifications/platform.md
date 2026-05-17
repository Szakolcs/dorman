# Platform and Cross-Cutting Specification

## Document Purpose

This specification translates platform requirements into implementable identity, authorization, audit, notification, and health behavior. It also defines **cross-module integration contracts** consumed by Administration, Doorman, Forum, and Chat. Requirement IDs refer to [platform requirements](../requirements/platform.md). Table-level detail is in [platform database](../database/platform.md).

---

## 1. Module Context

The platform layer is not a tenant-facing product module; it provides shared infrastructure:

- authentication and principal resolution for staff and tenants
- role-based authorization checks invoked by module services
- append-only audit logging for sensitive actions
- in-app notification enqueue and preference evaluation
- operational health reporting

### 1.1 Actor Model

- **Administrator**: audit search, role administration (where implemented), global policy
- **Staff roles** (office worker, doorman): authenticated via staff principal type
- **Tenant**: authenticated via tenant principal type with linked `Tenant` record
- **Platform Operator**: health endpoint consumer
- **System**: automated audit actors and notification workers

### 1.2 Module Boundaries

- **Administration and Office**: owns housing, inventory, maintenance, and official publication entities; emits audit on assignments and privileged publication
- **Doorman**: owns packages, guest visits, access logs, and loans; emits audit and package notification events
- **Forum**: owns feed engagement data; emits audit on moderation and privileged poll actions
- **Chat**: owns rooms and messages; consumes assignment events; emits optional message notification events

Out of scope for platform:

- business workflows and domain state machines owned by product modules
- forum feed ranking, poll vote rules, chat membership algorithms

---

## 2. Domain Model and Core Invariants

Entities are detailed in `docs/database/platform.md`. Conceptual core types:

- `User`, `Role`, `UserRole`
- `AuditEvent`
- `NotificationPreference` (per tenant, per category)
- `InAppNotification` (delivery row)
- module domain tables (referenced only by `target_type` / foreign keys in payloads)

### 2.1 Invariants (mandatory)

- **I-CC-01 Role assignment uniqueness**: at most one active `UserRole` row per (`user_id`, `role_id`) where `revoked_at IS NULL`.
- **I-CC-02 Audit append-only**: application code does not update or delete `AuditEvent` rows except documented admin tooling (if any).
- **I-CC-03 Authorization before mutation**: no module service performs a privileged write without a successful permission check.
- **I-CC-04 Preference gate**: non-critical notifications are suppressed when the tenant opted out of the category.
- **I-CC-05 Critical notices**: safety-category notifications bypass opt-out flags.

---

## 3. Capability Specifications

### 3.1 Authentication and Session Resolution

**Requirements:** `FR-CC-001`, `FR-CC-002`

#### Behavior

- Incoming requests resolve `Principal` = (`user_id`, `principal_type`, optional `tenant_id`, `role_names[]`).
- Staff routes require `principal_type = staff` and appropriate role.
- Tenant routes require `principal_type = tenant` and active tenant linkage.

#### Development stub (non-production)

- Header or cookie-based staff stub may set principal for local admin/forum/chat HTML workspaces; must not be enabled in production configuration.

---

### 3.2 Authorization Service

**Requirements:** `FR-CC-001`

#### Check contract

```
Authorize(principal, action, resource) -> allowed | denied
```

- `action` uses dotted names (`admin.room.assign`, `doorman.package.notify`, `forum.moderation.hide`).
- Denied checks return `authorization_denied` without partial writes.
- Sensitive denials may emit `audit.authorization.denied` when policy requires.

#### Baseline role matrix (illustrative)

| Action family | Administrator | Office | Doorman | Tenant |
|---------------|---------------|--------|---------|--------|
| Administration mutations | yes | delegated | no | no |
| Doorman desk | yes | read-only* | yes | no |
| Forum moderation | yes | yes | no | no |
| Tenant forum/chat | — | — | — | yes |

\*Exact office read scope is product policy; enforce in module-specific action lists.

---

### 3.3 Audit Writer

**Requirements:** `FR-CC-003`, `FR-CC-004`

#### Write contract

```
RecordAudit(actor_user_id, action, target_type, target_id, outcome, metadata?)
```

- `occurred_at` defaults to database timestamp.
- `outcome` ∈ {`success`, `failure`}.
- Modules call from the same transaction as the business mutation when failure must roll back audit+domain together.

#### Search contract (staff)

- Filters: `actor_user_id`, `action` prefix, `target_type`, date range
- Pagination required; default sort `occurred_at DESC`

#### Minimum cross-module action catalog

| Module | Example actions |
|--------|-----------------|
| Administration | `room_assignment.create`, `maintenance_ticket.status_change`, `forum_post.publish` |
| Doorman | `package.status_change`, `guest_access.denied`, `access.qr.validate` |
| Forum | `forum.moderation.hide`, `forum.poll.close_override` |
| Chat | `chat.moderation.override` (if implemented) |

---

### 3.4 Notifications

**Requirements:** `FR-CC-005`, `FR-CC-006`

#### Event intake

Modules publish logical events to a shared bus or direct service call:

| Event | Typical source | Category |
|-------|----------------|----------|
| `package.ready` | Doorman | package |
| `event.updated` | Forum / Administration | event |
| `poll.closed` | Forum | event |
| `chat.message.created` | Chat | chat |

#### Delivery pipeline

1. Resolve target tenant(s).
2. Load `NotificationPreference` for category.
3. If blocked and not critical, skip.
4. Insert `InAppNotification` row with payload JSON and deep link.
5. Record delivery status for support.

---

### 3.5 Health Endpoint

**Requirements:** `FR-CC-007`

#### Behavior

- `GET /health` (or `/ready` + `/live` split if adopted) returns JSON:
  - `status`: `ok` | `degraded` | `down`
  - `database`: connectivity probe result
  - optional `version` build metadata
- HTTP 503 when database unreachable and policy marks dependency required.

---

## 4. Interface-Level Contracts

### 4.1 Audit Query API (staff)

- pagination, filters as §3.3
- non-admin roles may be denied read access entirely or scoped to own actions per policy

### 4.2 Notification Preference API (tenant)

- `GET` / `PATCH` per category flags
- critical category flags are read-only `true`

### 4.3 Error Semantics

- `authorization_denied`
- `authentication_required`
- `preference_read_only` (critical category)
- `audit_write_failed` (internal; map to 500 for caller)

---

## 5. Non-Functional Specification

Mapped to `NFR-CC-001` through `NFR-CC-003`.

### 5.1 Security

- password hashes never returned in APIs
- health endpoint exposes no credentials

### 5.2 Reliability

- fail closed on audit+mutation coupling for compliance-critical administration and doorman actions

### 5.3 Performance

- index `AuditEvent(occurred_at DESC)`, `(actor_user_id, occurred_at DESC)`, `(action, occurred_at DESC)`

---

## 6. Cross-Module Interaction Specification

### 6.1 Administration Integration

- Staff sessions originate from platform `User` / `UserRole`.
- Assignment and publication services call `RecordAudit` on success and selected failures.
- Published content events may enqueue forum-visible notifications (`event.updated`).
- Assignment lifecycle events are consumed by chat (`RoomAssignmentCreated`, etc.) — chat module owns membership sync; platform only routes domain events if using an internal bus.

### 6.2 Doorman Integration

- Doorman principal requires `RoleDoorman` (or administrator override).
- Package notification flow creates doorman-owned `PackageNotification` rows and may call platform notification delivery for in-app channel.
- Guest denial and QR validation failures emit `AuditEvent` with `outcome = failure` when applicable.

### 6.3 Forum Integration

- Tenant forum actions bind `tenant_id` from principal.
- Moderation and poll override actions require staff roles; audited per forum spec.
- Optional `system_notice` feed items may reference doorman or platform notification payloads without duplicating package tables.

### 6.4 Chat Integration

- Tenant chat binds `tenant_id`; avatar storage uses platform blob helper.
- Flat membership is **not** platform-owned; platform does not write `ChatRoomMember`.
- `chat.message.created` events respect chat notification preference category.

### 6.5 Integration Matrix

| From → To | Mechanism | Data owned by source |
|-----------|-----------|----------------------|
| Administration → Forum | publication sync / `ForumPost` linkage | Activity, Event, official posts |
| Administration → Chat | assignment domain events | `RoomAssignment` |
| Doorman → Tenant UI | notifications / optional forum notice | `Package` |
| Forum → Platform | audit + notification events | engagement rows |
| Chat → Platform | audit (optional) + notification events | `ChatMessage` |
| All → Platform | `RecordAudit`, `Authorize` | — |

---

## 7. Compliance Mapping

| Specification section | Requirement mapping |
| --------------------- | ------------------- |
| Authentication and principals | FR-CC-001, FR-CC-002 |
| Authorization | FR-CC-001 |
| Audit | FR-CC-003, FR-CC-004 |
| Notifications | FR-CC-005, FR-CC-006 |
| Health | FR-CC-007 |
| Non-functional | NFR-CC-001 – NFR-CC-003 |

---

## 8. Implementation Notes

- Keep `AuditEvent` in a shared schema namespace; GORM model may live under administration package today but treat as platform-owned table.
- Prefer dotted `action` strings namespaced by module for searchability.
- Centralize `Authorize` in platform package; modules pass action constants.
- Notification workers should be idempotent on `(event_id, tenant_id)` if events can be retried.

---

## 9. Cross-References

- Features: [platform-features.md](../features/platform-features.md)
- User stories: [cross-cutting.md](../user-stories/cross-cutting.md)
- Use cases: [cross-cutting.md](../use-cases/cross-cutting.md)
- Database: [platform.md](../database/platform.md)
- Module specs: [administration-office.md](administration-office.md) §6, [doorman.md](doorman.md) §6, [forum.md](forum.md) §6, [chat.md](chat.md) §6
