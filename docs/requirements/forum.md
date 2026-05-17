# Forum Module Requirements

## Purpose

This document defines requirements for the Forum module: the tenant-facing communication surface for official dorm news, community activities and events, polls, and social engagement (comments and reactions). It aligns with [forum-features.md](../features/forum-features.md), [forum use cases](../use-cases/forum.md), and [forum user stories](../user-stories/forum.md).

## Scope

The module covers:

- Unified feed aggregating official and community content with sorting and pinning
- Post detail views for news, activities, events, and polls
- Community activity and event posts with schedule/location metadata and attendance intent
- Poll creation, eligibility enforcement, voting, and configurable result visibility
- Likes/reactions and comment threads with author ownership rules
- Staff moderation controls for community safety
- Browser-first HTML workspace at `/forum/view` for feed and form-based actions

The module shares boundaries with:

- **Administration and Office**: publishes official news, activities, and events that appear in the forum feed; remains source of truth for staff-owned publication lifecycle
- **Platform**: authentication, tenant/staff roles (including Student Consult Organizer), audit infrastructure, and optional notifications
- **Doorman / notifications**: optional surfacing of operational notifications in tenant feeds (doorman remains authoritative for package records)

## Actors and Roles

- **Tenant**: consumes feed, comments, reacts, votes in eligible polls, marks event/activity attendance intent
- **Student Consult Organizer**: creates and manages official events and polls (v1 may use staff-stub delegation per use case UC-FM-04)
- **Staff (Administrator / Office Worker)**: publishes official content via administration workflows; moderates community content
- **System**: enforces eligibility, vote limits, visibility rules, counters, and audit emission

## Functional Requirements

### Feed and Post Consumption

#### FR-FM-001 Unified Forum Feed

The system shall provide a centralized feed combining official and community posts.

Acceptance:

- feed includes published official news, announcements, activities, events, community posts, and open polls
- feed supports sort modes: newest first and pinned-first (pinned items ordered by pin priority then recency)
- draft and non-tenant-visible states are excluded from tenant feed

#### FR-FM-002 Post Detail and View Tracking

The system shall provide detail views for feed items and may record view activity for analytics.

Acceptance:

- detail exposes title, body, author identity, publication time, and type-specific metadata
- opening detail may append a view record linked to post and viewer (tenant/user) per product policy
- view tracking does not block read access when analytics write fails (degraded-safe)

### Community Activities and Events

#### FR-FM-003 Community Activity Posting

The system shall allow tenants to create and manage community activity posts.

Acceptance:

- post includes title, description, scheduled time, and location
- author can edit or archive own community activity while published
- archived posts are hidden from default feed but retain history for moderators

#### FR-FM-004 Official Event and Activity Visibility

The system shall display administration-published activities and events in the forum feed with participation metadata.

Acceptance:

- published staff content appears with organizer identity and schedule/location fields
- event posts expose capacity and registration deadline when configured
- postponement and cancellation update visible state on the feed item

#### FR-FM-005 Attendance Intent

The system shall allow tenants to mark planned attendance (going / not going) on activity and event posts.

Acceptance:

- one current intent per tenant per post; updates replace prior intent
- aggregate going / not-going counts are queryable for organizer dashboards
- intent changes are timestamped

#### FR-FM-006 Organizer Updates

The system shall allow authorized organizers to post updates on official events they manage.

Acceptance:

- updates are visible to tenants who marked going or not-going (or all viewers per product policy)
- updates preserve author and timestamp
- optional notification hook may inform interested tenants (integration with platform notifications)

### Polling

#### FR-FM-007 Poll Creation

The system shall support poll posts with configurable answer types and closing time.

Acceptance:

- poll supports single-choice and multiple-choice modes
- poll defines at least two options and a closing datetime
- only authorized roles (for example Student Consult Organizer or staff) can create official polls

#### FR-FM-008 Vote Eligibility and Limits

The system shall enforce vote eligibility and per-tenant vote limits according to poll configuration.

Acceptance:

- only eligible tenant accounts may submit votes
- single-choice polls accept at most one option per tenant
- multiple-choice polls accept at most one vote per option per tenant unless configured otherwise
- votes after close time are rejected

#### FR-FM-009 Poll Result Visibility

The system shall expose poll results according to a configured visibility mode.

Acceptance:

- `live` mode: aggregates update as votes arrive for eligible viewers
- `post_close` mode: aggregates hidden until closing time has passed
- results are persisted and exportable for organizers

### Engagement

#### FR-FM-010 Reactions

The system shall support adding and removing likes (or configured reaction type) on posts and comments.

Acceptance:

- one reaction per user per target (post or comment)
- reaction counts are reflected on feed and detail views
- removing a reaction decrements count consistently

#### FR-FM-011 Comment Threads

The system shall support comment threads on posts with author edit and delete rules.

Acceptance:

- tenants can add comments on posts they can view
- authors can edit or soft-delete their own comments within policy window
- deleted comments show placeholder or are hidden per moderation policy
- optional threaded replies via parent comment reference

#### FR-FM-012 Staff Moderation

The system shall provide staff moderation controls for community posts and comments.

Acceptance:

- staff can hide or remove violating community content
- moderation actions record actor, target, action, reason (optional), and timestamp
- moderated content is excluded from default tenant feed

### HTML Surface and API Parity

#### FR-FM-013 Forum HTML Workspace

The system shall expose a browser-first forum workspace under `/forum/view` for core tenant actions.

Acceptance:

- feed, post detail, comment form, poll vote form, and attendance intent form are reachable without a separate SPA requirement for v1
- successful form submissions persist domain state and refresh visible counters/results per policy
- same business rules apply to HTML forms and API mutations

#### FR-FM-014 Queryable Forum State

The system shall expose query APIs for feed, post detail, comments, poll state, and attendance aggregates consistent with HTML views.

Acceptance:

- list endpoints support pagination, filtering (for example kind, official vs community), and stable sorting
- mutation outcomes are readable from subsequent GET/list calls without stale counter drift under normal operation

### Security, Permissions, and Audit

#### FR-FM-015 Role-Based Access Control

The system shall enforce role-based permissions for poll creation, official updates, moderation, and privileged reads.

Acceptance:

- tenant mutations require authenticated tenant context
- organizer/staff actions require appropriate role grants
- authorization is enforced server-side; client role hints are non-authoritative

#### FR-FM-016 Sensitive Action Audit

The system shall emit audit records for moderation actions, poll configuration changes, and privileged content overrides.

Acceptance:

- audit records include actor, action, target, outcome, and timestamp
- domain tables (for example moderation log) provide queryable history in addition to platform audit where applicable

## Non-Functional Requirements

#### NFR-FM-001 Data Integrity

- vote, reaction, and attendance intent uniqueness constraints must be enforced transactionally
- foreign keys to tenant/user identities and administration-linked publications must remain consistent

#### NFR-FM-002 Traceability

- poll votes, moderation actions, and publication state changes must be reconstructable from persisted rows
- author attribution is preserved for posts, comments, votes, and reactions

#### NFR-FM-003 Performance

- feed and comment list endpoints must support pagination and index-backed filters (kind, published_at, pinned)
- aggregate counters (reactions, attendance, poll totals) should avoid unbounded full-table scans on hot paths

#### NFR-FM-004 Reliability

- validation failures return actionable errors without partial vote or intent state
- feed reads remain available when optional analytics writes fail

#### NFR-FM-005 Security

- tenants may not vote, react, or comment on behalf of other users
- moderated or draft content is not exposed through tenant APIs

## Business Rules

- A tenant may hold at most one active attendance intent per event/activity post.
- Poll votes cannot be changed after submission unless an explicit “vote change” policy is enabled; default is immutable vote per tenant per poll.
- Pinned posts appear before non-pinned items when sort mode is pinned-first.
- Community post authors may edit only their own posts; staff moderation overrides author visibility.
- Official publications authored via administration must not be editable by tenants in the forum module.
- Poll results in `post_close` mode must not leak option-level counts to tenants before close time.

## Out of Scope

- Semester tenant lifecycle, room assignment, and inventory (administration module)
- Package, guest, QR, and lending workflows (doorman module)
- Chat room membership and messaging (chat module)
- Full notification preference and delivery implementation (platform module; forum may emit events)

## Traceability Matrix

| Requirement | Source alignment |
|---|---|
| FR-FM-001, FR-FM-002 | FM-001, UC-FM-01, Unified Feed features |
| FR-FM-003, FR-FM-005 | FM-002, UC-FM-03, Community Activities features |
| FR-FM-004, FR-FM-006 | FM-003, administration publishing FR-AO-011–013 |
| FR-FM-007 – FR-FM-009 | FM-004, UC-FM-02, Polling features |
| FR-FM-010, FR-FM-011, FR-FM-012 | FM-005, Engagement features |
| FR-FM-013, FR-FM-014 | UC-FM-04, Forum HTML Surface features |
| FR-FM-015, FR-FM-016 | Platform RBAC and audit cross-cutting requirements |
