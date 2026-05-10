# Doorman Module Specification

## Document Purpose

This specification translates Doorman module requirements into implementable behavior, data constraints, workflow contracts, and integration boundaries. Requirement IDs refer to [doorman requirements](../requirements/doorman.md). Table-level detail is in [doorman database](../database/doorman.md).

---

## 1. Module Context

The Doorman module provides operational capabilities at the dormitory entry and front desk:

- package intake, status tracking, pickup, and optional tenant notification
- guest access desk: host linkage, visit windows, check-in/check-out, denial outside window
- QR validation for tenant entry with immutable access logging
- lending and return of keys and activity-linked accessories, including overdue surfacing

### 1.1 Actor Model

- **Doorman**: executes packages, guest handling, QR checks, lending
- **Tenant**: notification recipient; borrower of lendable items
- **Guest**: ephemeral visitor; identity bound to host tenant and visit slot
- **System**: state machine enforcement, time-window checks, audit emission

### 1.2 Module Boundaries

- **Administration**: source of truth for `Tenant`, `Room`, and `InventoryItem` definitions; doorman **reads** for linking and may reflect status updates for lendable stock according to policy
- **Platform**: `User`, `Role`, sessions, and `AuditEvent`
- **Tenant app / Forum / notifications**: optional surfaces for package notification delivery; doorman remains authoritative for package and lending records

Out of scope:

- semester lifecycle, room assignment rules, maintenance workflows (administration)
- forum discussions and social features

---

## 2. Domain Model and Core Invariants

Entities are detailed in `docs/database/doorman.md`. Conceptual core types:

- `Package` (or `Parcel`)
- `PackageNotification` (optional separate table or embedded log)
- `GuestAccess` / guest visit record with window
- `GuestAccessEvent` (check-in, check-out, denial) or unified access log
- `AccessEvent` (QR and/or entry gate events)
- `ItemLoan` / `AccessoryLoan` linking tenant, inventory or lendable item, dates
- platform `AuditEvent` for sensitive transitions

### 2.1 Invariants

- **I-DM-01 Package lifecycle**: status transitions follow the configured graph (no illegal rollback except explicit policy).
- **I-DM-02 Tenant link**: when tenant is set on a package, it must reference a valid `Tenant.id`.
- **I-DM-03 Guest window**: check-in is valid only if `now` ∈ [valid_from, valid_to] per stored rules; otherwise record denial, not check-in.
- **I-DM-04 QR uniqueness**: QR token or static identifier resolves to at most one active tenant context for entry validation.
- **I-DM-05 Lending single-flight**: one open loan per lendable unit unless inventory model defines quantity greater than one with separate stock rows.
- **I-DM-06 Traceability**: access and guest denial events include actor (doorman user id), timestamp, and outcome.

---

## 3. Capability Specifications

### 3.1 Package Intake and Pickup

**Requirements:** `FR-DM-001`, `FR-DM-002`, `FR-DM-003`

#### Behavior

- Doorman registers package with metadata and optional tenant match.
- System sets initial status to `received`.
- Optional transition to `notified` when notification is successfully queued.
- Pickup transitions to `picked_up` with handover timestamp and actor.

#### Validation

- Tenant reference optional; if lookup by name yields ambiguous results, UX must force disambiguation or leave tenant unset per UC-DM-01 alternate flow.

#### Audit (minimum)

- `package.create`
- `package.status_change`
- `package.pickup`

---

### 3.2 Tenant Notifications

**Requirements:** `FR-DM-004`, `FR-DM-005`

#### Behavior

- From a package in receivable state, doorman triggers notification only when tenant identity is unambiguous.

#### Validation

- No notification record marked successful when tenant link is missing or uncertain.

#### Integration contract

- Notification record stores `package_id`, `tenant_id`, `channel`, `created_at`, `status` (queued, sent, failed).

---

### 3.3 Guest Access Desk

**Requirements:** `FR-DM-006`, `FR-DM-007`

#### Behavior

- Register guest with host tenant and visit window before or at arrival.
- Check-in sets entry time; check-out sets exit time.
- Attempt to check in outside window: deny and log `guest.access_denied` with reason `outside_visit_window`.

#### State

- Visit record states may include: `scheduled`, `checked_in`, `checked_out`, `denied` (implementation choice; denial may be event-only).

---

### 3.4 QR Access Validation

**Requirements:** `FR-DM-008`, `FR-DM-009`

#### Behavior

- Scanner or manual entry resolves QR payload to a tenant credential or access token.
- Success: append `AccessEvent` with `granted`, tenant id, timestamp, optional device id.
- Failure: append `AccessEvent` with `denied`, reason code (`invalid`, `expired`, `revoked`).

#### Queries

- List API supports filters: `tenant_id`, date range, outcome.

---

### 3.5 Accessories and Key Lending

**Requirements:** `FR-DM-010`, `FR-DM-011`

#### Behavior

- Checkout creates open loan; return sets `returned_at` and clears checkout flag on inventory row or loan aggregate.
- Overdue when `expected_return_at` (if set) &lt; `now` and `returned_at` IS NULL.

#### Audit (minimum)

- `loan.checkout`
- `loan.return`
- `loan.overdue` (optional scheduled marker; at minimum list view computes overdue)

---

### 3.6 Security and Audit

**Requirements:** `FR-DM-012`, `FR-DM-013`

- All mutation endpoints require authenticated Doorman (or allowed role matrix from platform).
- Sensitive reads (guest PII, tenant details at scale) enforce same role checks.

---

## 4. Interface-Level Contracts

### 4.1 List APIs

Pending package queue, guest today’s list, access log, loan board: support pagination, filtering, stable sort (default newest first where applicable).

### 4.2 Mutation APIs

Package status change, guest check-in, QR validate, loan return: single transactional unit where multiple rows must stay consistent.

### 4.3 Error Semantics

- `validation_error`
- `tenant_not_found`
- `outside_visit_window`
- `package_invalid_transition`
- `loan_item_unavailable`
- `qr_invalid` / `qr_expired`
- `authorization_denied`
- `not_found`

---

## 5. Compliance Mapping

| Specification section        | Requirements        |
| ---------------------------- | ------------------- |
| Package intake and pickup    | FR-DM-001 – FR-DM-003 |
| Notifications                | FR-DM-004 – FR-DM-005 |
| Guest access                 | FR-DM-006 – FR-DM-007 |
| QR access                    | FR-DM-008 – FR-DM-009 |
| Lending                      | FR-DM-010 – FR-DM-011 |
| Security and audit           | FR-DM-012 – FR-DM-013 |

---

## 6. Cross-References

- Use cases: [doorman.md](../use-cases/doorman.md) (UC-DM-01–04)
- User stories: [doorman.md](../user-stories/doorman.md) (DM-001–005)
- Features checklist: [doorman-features.md](../features/doorman-features.md)
- Database: [doorman.md](../database/doorman.md)
