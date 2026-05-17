# Chat Module Requirements

## Purpose

This document defines requirements for the Chat module: tenant-facing real-time messaging for flat coordination, private conversations, and user-created group chats, including per-tenant chat profile personalization. It aligns with [chat-features.md](../features/chat-features.md), [chat use cases](../use-cases/chat.md), and [chat user stories](../user-stories/chat.md).

## Scope

The module covers:

- One system-managed flat room per flat with membership derived from active room assignments
- Direct (one-to-one) messaging between tenants with persistent history
- User-created group chats with membership management, naming, and avatars
- Chat profile fields: avatar, nickname (defaulting from tenant legal name), and bio
- Message send, chronological history, unread indicators, and near real-time delivery UX
- Browser-first HTML workspace at `/chat/view` with API parity

The module shares boundaries with:

- **Administration and Office**: source of truth for `Flat`, `Room`, `RoomAssignment`, and `Tenant` identity; assignment changes drive flat-room membership side effects
- **Platform**: authentication, tenant principal resolution, optional notification delivery, and audit infrastructure
- **Forum**: separate tenant communication surface; no shared post or poll entities

## Actors and Roles

- **Tenant**: participates in flat, direct, and group rooms; manages own chat profile; sends and reads messages
- **Staff (Administrator / Office Worker)**: may access privileged diagnostics or moderation hooks per platform policy; does not manually edit flat-room membership in v1
- **System**: provisions flat rooms, synchronizes derived membership, enforces room access, maintains read cursors, and applies delivery/idempotency rules

## Functional Requirements

### Flat Rooms

#### FR-CM-001 Flat Room Provisioning

The system shall maintain exactly one flat chat room per flat.

Acceptance:

- each `Flat` maps to at most one `ChatRoom` with `kind = flat`
- flat room is created when the flat is provisioned **or** lazily when the first active `RoomAssignment` exists for any room in that flat (MVP: **lazy creation on first qualifying assignment**)
- flat room metadata (title/icon) may default from flat label; tenants cannot delete the flat room

#### FR-CM-002 Flat Membership Synchronization

The system shall derive flat-room membership from active room assignments.

Acceptance:

- a tenant is a member of a flat room if and only if they have an active assignment (`ended_at IS NULL`) to any room whose `flat_id` matches the room’s flat
- membership add/remove follows assignment create, room change, and assignment end events without a separate manual join list for tenants
- non-admin tenants cannot add or remove flat-room members directly

### Direct and Group Conversations

#### FR-CM-003 Direct Messaging

The system shall support private one-to-one conversations between tenants.

Acceptance:

- a direct room can be opened from tenant profile or tenant search
- at most one direct room exists per unordered tenant pair (canonical pairing enforced in storage)
- message history loads in chronological order (`created_at ASC`, tie-break `id ASC`)
- only the two participant tenants may read or post in the direct room

#### FR-CM-004 Group Chat Lifecycle

The system shall allow tenants to create and manage group chats.

Acceptance:

- creator can set group name and avatar
- creator (or designated owner role) can add and remove participants who are valid tenants
- a member may leave a group; leaving sets membership inactive but preserves message history for remaining members
- removed members cannot post; historical messages they authored remain visible per product policy (default: visible)

#### FR-CM-005 Conversation List and Search

The system shall expose a tenant’s conversation list with recency and unread badges.

Acceptance:

- list includes flat room (when member), direct threads, and active group memberships
- each row shows display title, last message preview or timestamp, and unread count
- list supports search/filter by room title or participant nickname where configured

### Profile

#### FR-CM-006 Chat Profile Personalization

The system shall allow each tenant to customize chat-visible profile fields.

Acceptance:

- fields include avatar, nickname, and bio
- default nickname is the tenant’s full legal name from the tenant record until overridden
- avatar upload validates allowed MIME types and maximum size
- profile changes appear in room headers, member lists, and message author labels without requiring message backfill

### Messaging

#### FR-CM-007 Message Send and History

The system shall persist chat messages per room with author attribution.

Acceptance:

- authorized members can post text messages (attachments out of scope for v1 unless explicitly added)
- message history is paginated newest-first or oldest-first per API contract; default tenant timeline is chronological ascending
- authors may edit or soft-delete their own messages within policy window (optional v1: edit/delete deferred; if deferred, document as post-MVP)
- server rejects posts from non-members and from inactive memberships

#### FR-CM-008 Read State and Unread Counts

The system shall track per-member read position and surface unread counts.

Acceptance:

- opening a conversation updates the member’s read cursor to the latest visible message
- unread count for a room is the number of messages after the member’s read cursor from other authors (or all messages after cursor per product policy)
- read state updates are visible in conversation list badges and per-room indicators

#### FR-CM-009 Near Real-Time Delivery

The system shall deliver new messages to connected clients with minimal latency.

Acceptance:

- clients receive new messages without full-page reload (polling, SSE, or WebSocket — transport is implementation choice)
- transient send failures expose a retryable client state; duplicate retries with the same client idempotency key do not create duplicate persisted messages
- per-conversation delivery/read indicators update when peers read messages (minimum: read cursor advancement; optional: per-message receipt rows)

### HTML Surface and API Parity

#### FR-CM-010 Chat HTML Workspace

The system shall expose a browser-first chat workspace under `/chat/view`.

Acceptance:

- workspace lists conversations and opens room timelines with a send form
- tenants can create group chats and open direct threads as allowed by policy
- profile edit page updates nickname, avatar, and bio via form posts
- staff-stub or test actors may use the same routes in non-production configurations (UC-CM-04)

#### FR-CM-011 Queryable Chat APIs

The system shall expose JSON/query APIs consistent with HTML views.

Acceptance:

- endpoints cover conversation list, room detail, message pages, membership changes (group), profile read/update, and send message
- HTML handlers and API handlers invoke the same application services
- list endpoints support pagination cursors and stable ordering

### Security and Audit

#### FR-CM-012 Room Access Control

The system shall enforce tenant-scoped access to rooms and messages.

Acceptance:

- every read and write requires authenticated tenant context
- principals may access only rooms where they have an active `ChatRoomMember` row (flat membership included)
- authorization is enforced server-side

#### FR-CM-013 Sensitive Action Audit

The system shall emit audit records for privileged chat operations when supported.

Acceptance:

- minimum audited actions: staff moderation override (if implemented), forced membership changes (if any), and profile moderation overrides
- audit rows include actor, action, target, outcome, and timestamp

## Non-Functional Requirements

#### NFR-CM-001 Data Integrity

- enforce unique flat room per `flat_id`, unique direct room per tenant pair, and unique active membership per (`room_id`, `tenant_id`) where applicable
- flat membership sync and message insert for a single send must be transactionally consistent with membership checks

#### NFR-CM-002 Traceability

- messages retain `author_tenant_id`, `created_at`, and room linkage for history reconstruction
- assignment-driven membership changes are reconstructable from `RoomAssignment` timelines plus optional sync log

#### NFR-CM-003 Performance

- message list and conversation list endpoints use index-backed pagination (`room_id`, `created_at`)
- unread counts should avoid full-table scans on hot paths (maintain read cursor or bounded count strategy)

#### NFR-CM-004 Reliability

- idempotent message send via client-supplied key prevents duplicates on retry
- read cursor updates are safe under concurrent opens (last-read advances monotonically)

#### NFR-CM-005 Security

- tenants cannot read or post in rooms where they are not members
- avatar upload endpoints validate type/size and store outside web root per platform file policy

## Business Rules

- Flat-room membership is **derived only** from active assignments; it is not a manually curated roommate list in v1.
- Exactly one flat `ChatRoom` exists per `Flat`.
- Exactly one direct `ChatRoom` exists per pair of tenants (order-independent).
- Group creators hold an owner role that can add/remove members unless ownership is transferred (transfer optional in v1).
- Leaving a group deactivates membership but does not delete the room or its message history.
- Default chat nickname mirrors tenant legal name until the tenant sets a nickname.
- Unread badges clear when the tenant opens the conversation and the read cursor advances.

## Out of Scope

- Forum posts, polls, reactions, and moderation queues (forum module)
- Package, guest, QR, and lending workflows (doorman module)
- Room assignment and inventory administration (administration module; chat consumes side effects only)
- End-to-end encryption, message reactions, threads, and file attachments (unless added by later requirement)
- Full notification preference and push delivery implementation (platform module; chat may emit events)

## Traceability Matrix

| Requirement | Source alignment |
|---|---|
| FR-CM-001, FR-CM-002 | CM-001, UC-CM-01, Flat Room Automation features |
| FR-CM-003, FR-CM-005, FR-CM-008 | CM-002, UC-CM-02, Direct Messaging features |
| FR-CM-004 | CM-003, UC-CM-03, Group Chats features |
| FR-CM-006 | CM-004, Profile Features |
| FR-CM-007, FR-CM-009 | CM-005, Messaging UX features |
| FR-CM-010, FR-CM-011 | UC-CM-04, Chat HTML Surface features |
| FR-CM-012, FR-CM-013 | Platform RBAC and audit cross-cutting requirements |
