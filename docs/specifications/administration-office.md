# Administration and Office Module Specification

## Document Purpose

This specification translates Administration and Office module requirements into implementable system behavior, data constraints, workflow contracts, and integration boundaries.

---

## 1. Module Context

The Administration and Office module provides staff-facing operations for dormitory management:

- tenant lifecycle administration
- room allocation and occupancy control
- inventory lifecycle tracking
- maintenance intake, approval, assignment, and resolution
- staff operational scheduling
- official communication publishing (news, activities, events)

### 1.1 Actor Model

- **Administrator**: full operational scope, approval and override actions
- **Office Worker**: day-to-day operational actor for tenant/room/content/ticket workflows
- **Director (optional role profile)**: supervisory maintenance and inventory oversight
- **System**: validator, enforcer of invariants, and audit emitter

### 1.2 Module Boundaries

- **Forum module**: receives published staff content for tenant consumption
- **Chat module**: receives side effects from room assignment changes
- **Doorman module**: consumes shared tenant/room/inventory context where needed
- **Platform layer**: provides authn/authz and shared audit substrate

Out of scope for this module:

- tenant social interactions (comments, reactions as social features, polls)
- chat CRUD and message lifecycle
- package/guest/lending operations owned by doorman workflows

---

## 2. Domain Model and Core Invariants

This section captures required behavior over entities defined in `docs/database/administration.md`.

### 2.1 Core Entities

- `Tenant`
- `Flat`
- `Room`
- `RoomAssignment`
- `InventoryItem`
- `MaintenanceTicket`
- `TicketStatusChange`
- `OperationalJob`
- `Activity`
- `Event`
- `ForumPost` linkage for official publications
- platform-provided `User` / role mapping and `AuditEvent`

### 2.2 Invariants (mandatory)

- **I-01 Active assignment uniqueness**: one tenant has at most one active `RoomAssignment`.
- **I-02 Capacity safety**: active assignments per room must not exceed `Room.capacity`.
- **I-03 Temporal assignment history**: assignment change is modeled as ending old active row and creating a new row.
- **I-04 Ticket lifecycle integrity**: ticket status transitions must follow allowed state edges.
- **I-05 Traceability**: all critical staff writes preserve actor and timestamp context.
- **I-06 Publication identity**: published news/activity/event keeps author and publication time.

---

## 3. Capability Specifications

Each capability references requirement IDs from `docs/requirements/administration-office.md`.

### 3.1 Tenant Lifecycle Administration

**Requirements:** `FR-AO-001`, `FR-AO-002`

#### 3.1.1 Behavior

- Staff can mark tenants inactive when they leave.
- Staff can reactivate returning tenants for next semester.
- Staff can register newly arriving tenants.
- Tenant directory supports status-based filtering and drill-down details.

#### 3.1.2 Validation Rules

- Student status checks must pass before lifecycle mutation.
- Lifecycle mutation must include acting staff identity.

#### 3.1.3 Audit Events (minimum)

- `tenant.activate`
- `tenant.deactivate`
- `tenant.register`
- `tenant.update`

---

### 3.2 Room and Occupancy Management

**Requirements:** `FR-AO-003`, `FR-AO-004`, `FR-AO-005`

#### 3.2.1 Behavior

- Room directory shows occupancy and inventory status.
- Individual tenant assignment/reassignment is supported.
- Bulk allocation planning can be generated, edited, and approved pre-semester.

#### 3.2.2 Validation Rules

- Assignment must fail with a capacity error if room is full.
- Assignment operation must be transactional over assignment history and occupancy views.
- Bulk allocation approval must apply as a consistent batch.

#### 3.2.3 Consistency Rules

- Assignment write path enforces `I-01` and `I-02`.
- Reassignment path:
  1. close previous active assignment (`ended_at`)
  2. insert new active assignment (`ended_at = NULL`)
  3. emit audit event

#### 3.2.4 Side Effects

- Flat/room membership changes can trigger downstream chat membership synchronization.

---

### 3.3 Inventory Management

**Requirements:** `FR-AO-006`, `FR-AO-007`

#### 3.3.1 Behavior

- Track room-scoped and shared-area inventory items.
- Maintain lifecycle statuses (active/in use/withdrawn/destroyed or equivalent configured statuses).
- Generate inventory status report for physical reconciliation.

#### 3.3.2 Validation Rules

- Inventory item must have an unambiguous location context.
- Lifecycle transitions must preserve historical traceability.

#### 3.3.3 Audit Events (minimum)

- `inventory.create`
- `inventory.update`
- `inventory.status_change`
- `inventory.audit_reconciliation`

---

### 3.4 Maintenance Management

**Requirements:** `FR-AO-008`, `FR-AO-009`

#### 3.4.1 Behavior

- Create staff-originated maintenance tickets.
- Review and approve tenant-originated tickets.
- Assign tickets to worker/team and due date.
- Advance ticket through lifecycle until resolved/closed.

#### 3.4.2 Ticket State Model

Baseline states (aligned to current docs and implementation flexibility):

- `reported`
- `duplicate`
- `in_progress`
- `halted`
- `resolved`
- `closed`

Allowed transitions (minimum contract):

- `reported -> in_progress`
- `reported -> duplicate`
- `in_progress -> halted`
- `halted -> in_progress`
- `halted -> closed`
- `in_progress -> resolved`
- `resolved -> closed`

Optional administrative override transitions are allowed only for authorized roles and must emit elevated audit context.

#### 3.4.3 Lifecycle Logging

- Every status transition appends one `TicketStatusChange` row with actor and timestamp.
- Ticket detail view must expose ordered transition history.

---

### 3.5 Operational Scheduling

**Requirements:** `FR-AO-010`

#### 3.5.1 Behavior

- Staff can create operational jobs with assignee, priority, and time window.
- Jobs are viewable in list and calendar representations.
- Scheduler highlights overlap conflicts before save.

#### 3.5.2 Conflict Contract

- Conflict check is computed on `[starts_at, ends_at)` by assignee.
- Hard block or warning-only mode is configurable; default is warning with explicit user confirmation required for save.

---

### 3.6 Official Publishing (News, Activities, Events)

**Requirements:** `FR-AO-011`, `FR-AO-012`, `FR-AO-013`

#### 3.6.1 Behavior

- Staff can draft and publish official news items.
- Staff can create/publish activities with capacity and optional accessory references.
- Staff can create/publish events with tenant reaction/participation intent support.

#### 3.6.2 Publication States

Minimum state model:

- `draft`
- `published`
- `archived`
- `canceled` (events where applicable)
- `postponed` (events where applicable)

#### 3.6.3 Integration Contract (Forum)

- Published content must be represented in forum-visible feed records.
- Publication writes must preserve staff author identity and publish timestamp.

---

### 3.7 Security and Audit

**Requirements:** `FR-AO-014`, `FR-AO-015`

#### 3.7.1 Authorization

- Every mutation endpoint enforces staff-role permission checks.
- Server-side authorization is authoritative; client role hints are non-authoritative.

#### 3.7.2 Audit Contract

Sensitive mutations must emit `AuditEvent` containing:

- actor identity
- action identifier
- target entity reference
- operation result
- timestamp

---

## 4. Interface-Level Contracts

This section defines behavioral contracts independent of transport protocol details.

### 4.1 List APIs

Lists for tenants, rooms, tickets, jobs, inventory must support:

- pagination
- filtering
- deterministic sorting

### 4.2 Mutation APIs

Mutations that involve multiple entities (for example reassignment) must:

- execute transactionally
- return explicit validation/conflict errors
- avoid partial persistence on failure

### 4.3 Error Semantics

Minimum domain error categories:

- `validation_error`
- `capacity_conflict`
- `state_transition_invalid`
- `authorization_denied`
- `not_found`
- `concurrency_conflict`

---

## 5. Non-Functional Specification

Mapped to `NFR-AO-001` through `NFR-AO-005`.

### 5.1 Data Integrity and Reliability

- enforce transactional consistency for assignment and lifecycle-critical writes
- maintain foreign-key and domain-constraint integrity
- reject invalid states with actionable errors

### 5.2 Traceability and Auditability

- reconstruct business-critical state transitions from persisted domain history
- preserve actor context for sensitive actions

### 5.3 Performance

- list operations must be index-supported for frequent operational filters
- pagination defaults should prevent unbounded list retrieval

### 5.4 Security

- all staff operations require authenticated principal context
- role checks enforced on every privileged action

---

## 6. Cross-Module Interaction Specification

### 6.1 Forum Integration

- Official publications become tenant-visible forum feed entries.
- Administration remains source of truth for staff-owned publication lifecycle.

### 6.2 Chat Integration

- Room assignment changes are authoritative source for flat-membership side effects.
- Administration does not directly manage chat room lifecycle.

### 6.3 Doorman Integration

- Doorman workflows may read room/tenant/inventory context.
- Ownership of doorman entities remains external to this module.

---

## 7. Compliance Mapping


| Specification section           | Requirement mapping             |
| ------------------------------- | ------------------------------- |
| Tenant Lifecycle Administration | FR-AO-001, FR-AO-002            |
| Room and Occupancy Management   | FR-AO-003, FR-AO-004, FR-AO-005 |
| Inventory Management            | FR-AO-006, FR-AO-007            |
| Maintenance Management          | FR-AO-008, FR-AO-009            |
| Operational Scheduling          | FR-AO-010                       |
| Official Publishing             | FR-AO-011, FR-AO-012, FR-AO-013 |
| Security and Audit              | FR-AO-014, FR-AO-015            |
| Non-Functional Specification    | NFR-AO-001 to NFR-AO-005        |


---

## 8. Implementation Notes

- Use `RoomAssignment` as temporal assignment record rather than duplicating into a separate history table.
- Prefer append-only lifecycle records for maintenance transitions (`TicketStatusChange`) and sensitive domain events (`AuditEvent`).
- Keep state machines explicit in service layer to avoid invalid transition drift.
- Preserve flexible naming alignment between documentation and migration naming conventions (snake_case in SQL/GORM layer).

