# Administration and Office Module Requirements

## Purpose

This document defines requirements for the Administration and Office module used by administrators and office workers to manage tenants, rooms, inventory, maintenance operations, scheduling, and official tenant-facing communication.

## Scope

The module covers:

- Tenant lifecycle administration for semesters
- Room allocation and capacity-safe assignment
- Inventory lifecycle and accountability
- Maintenance ticket approval and execution lifecycle
- Operational staff job scheduling
- Publishing official news, activities, and events for tenants
- Auditability of sensitive operational changes

The module shares boundaries with:

- Forum module for published communication visibility
- Chat module for flat membership side effects driven by room assignment
- Doorman module for package/access/lending operations that may reference shared entities
- Platform module for authentication, role management, and audit infrastructure

## Actors and Roles

- **Administrator**: Full module access, including approvals and overrides
- **Office Worker**: Operational CRUD within delegated module permissions
- **Director**: Optional operational supervisor role for maintenance/inventory workflows
- **System**: Enforces constraints, records audit history, and publishes integration side effects

## Functional Requirements

### Tenant Management

#### FR-AO-001 Semester Tenant Lifecycle

The system shall support semester lifecycle updates for tenants, including:

- mark leaving tenants inactive
- carry forward returning tenants to active state
- register newly arriving tenants

Acceptance:

- lifecycle actions are validated against student status constraints
- each change is attributable to an acting staff user and timestamped

#### FR-AO-002 Tenant Directory and Detail

The system shall provide searchable tenant lists and detail views with status indicators required for assignment and administration workflows.

Acceptance:

- list supports filtering by active/inactive status
- tenant detail includes assignment and relevant administration attributes

### Room and Capacity Management

#### FR-AO-003 Room and Occupancy Overview

The system shall provide room list and detail views showing room number, capacity, occupancy, and inventory status.

Acceptance:

- list can be filtered by occupied, available, and maintenance-needed states
- room details expose assigned tenants and linked inventory items

#### FR-AO-004 Individual Room Assignment

The system shall allow assigning and reassigning one tenant to one room while enforcing occupancy rules.

Acceptance:

- assignment is rejected when target room capacity would be exceeded
- assignment updates active occupancy immediately
- assignment and reassignment are historically auditable

#### FR-AO-005 Bulk Room Allocation Planning

The system shall support generating and approving planned room allocations for all active tenants during pre-semester setup.

Acceptance:

- allocation proposal can be generated from active tenant and room inputs
- staff can adjust proposal (for example tenant swaps) before approval
- approval persists resulting assignments as auditable records

### Inventory Management

#### FR-AO-006 Inventory Registration and Tracking

The system shall support registration and lifecycle tracking of room-level and shared-area inventory items.

Acceptance:

- items track availability and condition status
- items can be associated with room or shared-area location context
- item lifecycle changes (for example withdrawn/destroyed) are auditable

#### FR-AO-007 Inventory Reporting and Audit Support

The system shall provide inventory reports that support reconciliation with physical audits.

Acceptance:

- report includes current status and location context for all tracked items
- updates after physical audit persist change history

### Maintenance Management

#### FR-AO-008 Maintenance Ticket Creation and Approval

The system shall support ticket intake from office staff and approval workflow for tenant-submitted tickets.

Acceptance:

- ticket captures category, severity, impact, location, and description
- tenant-submitted tickets can be listed and filtered by approval state
- approved tickets can be assigned to staff/team with due date

#### FR-AO-009 Maintenance Lifecycle Tracking

The system shall support end-to-end maintenance lifecycle transitions with actor/time history.

Acceptance:

- workflow supports statuses from reported/open to resolved/closed
- each status transition records actor and timestamp
- lifecycle history is queryable without reconstructing from raw logs

### Operational Scheduling

#### FR-AO-010 Staff Job Scheduling

The system shall support scheduling operational jobs with assignee, priority, and time window.

Acceptance:

- schedule can be viewed in list and calendar forms
- filters include assignee and date
- overlapping assignments are highlighted before save

### Official Communication and Activities

#### FR-AO-011 Official News Publishing

The system shall support drafting and publishing official dorm news/articles for tenant visibility.

Acceptance:

- content supports title, body, tags, and publish date
- draft and published states are supported
- published entries appear in tenant forum feed

#### FR-AO-012 Activity Publishing

The system shall support creating and publishing dorm activities with capacity and optional accessory linkage.

Acceptance:

- activity includes title, location, schedule, and capacity
- activity can reference required accessories/keys where needed
- published activity appears in tenant-facing forum/activity surfaces

#### FR-AO-013 Event Publishing

The system shall support creating and publishing dorm events with tenant participation interaction.

Acceptance:

- event includes title, location, schedule, and capacity
- tenants can react or indicate participation intent
- events can be postponed or canceled with updated visibility state

### Security, Permissions, and Audit

#### FR-AO-014 Role-Based Access Control

The system shall enforce role-based permissions for all administration-office operations.

Acceptance:

- only authorized staff roles can perform mutation operations
- permission checks apply at API and UI action points

#### FR-AO-015 Sensitive Action Audit

The system shall emit audit records for sensitive module actions, including assignment, ticket overrides, and privileged content publication.

Acceptance:

- audit records include actor, action, target, outcome, and timestamp
- records are retained for compliance and operational investigation

## Non-Functional Requirements

#### NFR-AO-001 Data Integrity

- room capacity and active assignment constraints must be enforced transactionally
- foreign-key integrity must be maintained across tenant, room, ticket, inventory, and staff references

#### NFR-AO-002 Traceability

- business-critical state transitions must be historically reconstructable
- records must preserve actor attribution for operational accountability

#### NFR-AO-003 Performance

- list endpoints for rooms, tenants, tickets, and jobs should support pagination, filtering, and stable sorting
- common list queries should be indexed for operationally responsive UI use

#### NFR-AO-004 Availability and Reliability

- write operations must fail safely and return actionable validation errors
- no partial updates are allowed for multi-entity operations such as assignments

#### NFR-AO-005 Security

- all staff operations require authenticated principal context
- authorization decisions must not rely on client-provided role hints alone

## Business Rules

- A tenant can have at most one active room assignment at a time.
- Active assignments in a room cannot exceed room capacity.
- Maintenance status progression must follow the configured lifecycle.
- Published communication must preserve author identity and publication timestamps.
- Bulk allocation approval is restricted to authorized staff roles during pre-semester flow.

## Out of Scope

- Tenant social interactions (comments, likes, polls) owned by forum module
- Chat room and membership CRUD owned by chat module (assignment changes only drive side effects)
- Doorman-specific package, guest access, and lending workflows

## Traceability Matrix

| Requirement | Source alignment |
|---|---|
| FR-AO-001, FR-AO-002 | User stories AO-000, use case UC-AO-00, features Tenant Management |
| FR-AO-003, FR-AO-004 | User stories AO-001, AO-002, use case UC-AO-01, features Room and Capacity Management |
| FR-AO-005 | Use case UC-AO-02, database assignment temporal model guidance |
| FR-AO-006, FR-AO-007 | Use case UC-AO-04, features Inventory Management |
| FR-AO-008, FR-AO-009 | User stories AO-003, maintenance approval story, use case UC-AO-03, features Maintenance Management |
| FR-AO-010 | User story AO-005, features Operational Scheduling |
| FR-AO-011, FR-AO-012, FR-AO-013 | User stories AO-006, AO-007, AO-008, use cases UC-AO-05 to UC-AO-07, features News and Activity Publishing |
| FR-AO-014, FR-AO-015 | Database RBAC/audit dependencies, platform cross-cutting requirements |
