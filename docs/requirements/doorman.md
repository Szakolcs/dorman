# Doorman Module Requirements

## Purpose

This document defines requirements for the Doorman module: front-desk operations for package handling, tenant notifications around deliveries, guest access control, QR-based dorm entry validation, and lending of shared keys and activity accessories. It aligns with [doorman-features.md](../features/doorman-features.md), [doorman use cases](../use-cases/doorman.md), and [doorman user stories](../user-stories/doorman.md).

## Scope

The module covers:

- Package intake and pickup with a clear status lifecycle
- Tenant-facing notification when a package is ready (where tenant identity is certain)
- Guest access: registration, validity windows, check-in and check-out, denial when outside the window
- QR-based validation for authorized tenant entry and immutable access event logs
- Checkout and return of keys and activity accessories, including overdue visibility

The module shares boundaries with:

- **Administration**: owns `Tenant`, room assignment, and master inventory definitions; doorman consumes that context for linking packages, guests, and loans
- **Platform**: authentication, roles (Doorman), and audit infrastructure
- **Forum / notifications**: tenant-visible delivery or system notifications may surface in the tenant app feed per product integration

## Actors and Roles

- **Doorman**: primary operator for packages, guest desk, QR validation, and lending workflows
- **Tenant**: receives package-related notifications; may borrow/return items subject to rules
- **Guest**: physical visitor; identity recorded against a host tenant and time window
- **System**: enforces visit windows, package rules, generates notifications, and records audit events

## Functional Requirements

### Package Intake and Pickup

#### FR-DM-001 Package Registration

The system shall allow doorman staff to register an incoming package and associate it with a tenant when identification is possible.

Acceptance:

- package record stores linkage to tenant (or explicit “unidentified” handling), arrival time, and descriptive metadata (for example label name, package type)
- initial status reflects successful intake (for example received)

#### FR-DM-002 Package Lifecycle

The system shall support a defined package status lifecycle from intake through pickup.

Acceptance:

- statuses include at minimum: received, notified (optional), picked up
- transitions are valid only along allowed edges; terminal state pickup is recorded with timestamp

#### FR-DM-003 Pickup Confirmation

The system shall support confirming handover when the tenant collects a package.

Acceptance:

- pickup records actor (doorman user), time, and resulting status
- picked-up packages leave the pending pickup queue

### Tenant Notifications (Packages)

#### FR-DM-004 Package Notification

The system shall support triggering an in-app notification from the package record when a matching tenant is identified with sufficient certainty.

Acceptance:

- notification can be initiated from the package workflow using tenant identity derived from full name or other agreed matching rules
- if the tenant cannot be identified reliably, no automated notification is sent (operator may still complete intake)
- notification log stores channel, timestamp, and link to package

#### FR-DM-005 Tenant Visibility

The tenant shall be able to see relevant package or delivery notifications in their application feed when the product integrates notifications with the tenant app.

Acceptance:

- notification is associated with the package and tenant record where applicable

### Guest Access

#### FR-DM-006 Guest Access Record

The system shall support guest visits tied to a host tenant and a declared visit time window.

Acceptance:

- record includes host tenant, guest identity, and valid-from / valid-to (or equivalent window)
- check-in and check-out timestamps are recorded when the doorman completes those actions

#### FR-DM-007 Guest Entry Validation

The system shall deny entry when the current time is outside the approved visit window and shall log the denial.

Acceptance:

- within window: check-in allowed per workflow rules
- outside window: access denied; event is auditable

### QR Access Validation

#### FR-DM-008 QR Verification

The system shall validate tenant QR identifiers at entry and record the outcome.

Acceptance:

- valid code: entry allowed and access event created with timestamp
- invalid or expired code: access denied and event logged
- doorman workspace receives tenant details when validation succeeds

#### FR-DM-009 Access Log Queries

The system shall support listing and filtering access events (for example by tenant and date range).

Acceptance:

- filters support operational review and auditing

### Accessories and Key Lending

#### FR-DM-010 Lending Record

The system shall support checking out and returning inventory items (keys, activity accessories) to tenants with timestamps.

Acceptance:

- lending captures item reference, tenant, checkout time, and expected return when applicable
- return closes the lending record and restores availability for lendable items

#### FR-DM-011 Overdue and Availability

The system shall surface overdue loans and support a view of availability for shared equipment.

Acceptance:

- overdue items are identifiable in a dedicated operational view
- product may trigger reminder hooks (email/in-app) per integration; core requirement is visible flagging

### Security, Permissions, and Audit

#### FR-DM-012 Role-Based Access Control

The system shall restrict doorman operations to authorized principals (for example platform role Doorman and delegated staff policies).

Acceptance:

- mutations require authenticated Doorman (or allowed staff) context
- authorization is enforced server-side

#### FR-DM-013 Sensitive Action Audit

The system shall emit audit records for security-sensitive actions: guest denial, QR denial, package status changes, and lending close.

Acceptance:

- audit records include actor, action, target, outcome, and timestamp

## Non-Functional Requirements

#### NFR-DM-001 Data Integrity

- foreign keys to `Tenant`, `User`, and shared inventory references must remain consistent
- lending and package updates must not leave partial orphan state without remediation paths

#### NFR-DM-002 Traceability

- access and guest decisions must be reconstructable from persisted domain rows and audit where applicable

#### NFR-DM-003 Performance

- queue and list views (packages pending pickup, overdue loans, recent access events) should support pagination and indexed filters

#### NFR-DM-004 Reliability

- validation and denial paths return actionable errors without silent failure

#### NFR-DM-005 Security

- QR validation must not expose tenant PII beyond what the doorman role is allowed to see
- guest data is visible only to authorized roles

## Business Rules

- A package in “picked up” state cannot return to “received” without an explicit administrative override policy if ever allowed.
- Guest entry outside the declared window is denied.
- Lending availability must reflect whether an item is currently on loan (single checkout constraint per lendable unit unless product defines pooled quantity).

## Out of Scope

- Room allocation and tenant lifecycle (administration module)
- Forum social threads and moderation (forum module)
- Chat room membership (chat module)

## Traceability Matrix

| Requirement | Source alignment |
|---|---|
| FR-DM-001 – FR-DM-003 | DM-001, UC-DM-01, Package intake and pickup features |
| FR-DM-004 – FR-DM-005 | DM-002, UC-DM-01 alternate flow, notification features |
| FR-DM-006 – FR-DM-007 | DM-003, UC-DM-02 |
| FR-DM-008 – FR-DM-009 | DM-004, UC-DM-04 |
| FR-DM-010 – FR-DM-011 | DM-005, UC-DM-03, Accessories features |
| FR-DM-012 – FR-DM-013 | Platform RBAC, audit cross-cutting requirements |
