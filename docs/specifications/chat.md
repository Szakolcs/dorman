# Chat Module Specification

## Document Purpose

This specification translates Chat module requirements into implementable system behavior, data constraints, workflow contracts, and integration boundaries. Requirement IDs refer to [chat requirements](../requirements/chat.md). Table-level detail is in [chat database](../database/chat.md).

---

## 1. Module Context

The Chat module is the tenant-facing messaging layer:

- one automatic flat channel per flat, membership synced from housing assignments
- direct messages between tenants
- user-created group chats with owner-controlled membership
- per-tenant chat profile (avatar, nickname, bio)
- near real-time timelines with unread/read state
- HTML-first workspace at `/chat/view` with API parity

### 1.1 Actor Model

- **Tenant**: read/send messages, manage group membership (where permitted), edit own chat profile
- **Staff**: optional diagnostics, moderation, or staff-stub access in non-production (UC-CM-04); does not curate flat membership in v1
- **System**: provisions flat rooms, syncs derived membership on assignment events, enforces access, maintains read cursors

### 1.2 Module Boundaries

- **Administration and Office**: authoritative for `Flat`, `Room`, `RoomAssignment`, `Tenant` legal identity; chat listens for assignment lifecycle events
- **Platform**: `User` / tenant authentication, file storage for avatars, optional `AuditEvent` and notification workers
- **Forum**: separate feed and engagement model; no shared tables

Out of scope:

- forum posts, polls, comments, and feed moderation
- doorman operational records
- housing assignment algorithms and inventory CRUD

---

## 2. Domain Model and Core Invariants

Entities are detailed in `docs/database/chat.md`. Conceptual core types:

- `ChatRoom` (`kind`: `flat`, `direct`, `group`)
- `ChatRoomMember` (membership, roles, read cursor, leave state)
- `ChatMessage`
- `ChatTenantProfile`
- administration `Flat`, `Room`, `RoomAssignment`, `Tenant`
- platform `User`, optional `AuditEvent`

### 2.1 Invariants (mandatory)

- **I-CM-01 One flat room per flat**: at most one `ChatRoom` with `kind = flat` per `flat_id`.
- **I-CM-02 Derived flat membership**: tenant `T` is an active flat-room member iff ∃ active `RoomAssignment` for `T` in any `Room` with `room.flat_id = chat_room.flat_id`.
- **I-CM-03 Direct uniqueness**: at most one direct room per unordered tenant pair (`tenant_low_id`, `tenant_high_id`).
- **I-CM-04 Posting membership**: only tenants with active `ChatRoomMember` (`left_at IS NULL`) may read/post.
- **I-CM-05 Message ordering**: messages within a room are strictly ordered by (`created_at`, `id`).
- **I-CM-06 Read monotonicity**: `last_read_message_id` (or equivalent cursor) only advances forward for a given member.
- **I-CM-07 Idempotent send**: duplicate client idempotency keys map to one persisted `ChatMessage`.
- **I-CM-08 Profile default**: unset nickname displays tenant legal full name from administration `Tenant` record.

---

## 3. Capability Specifications

### 3.1 Flat Room Provisioning and Membership

**Requirements:** `FR-CM-001`, `FR-CM-002`

#### Provisioning (MVP)

- **Lazy create:** on first event that implies flat occupancy (active assignment created for any room in flat `F`), upsert `ChatRoom{ kind: flat, flat_id: F }` if missing.
- Room title defaults to flat label (for example `Flat 3B`); not user-deletable.

#### Membership sync

Triggered on:

- assignment create (new active row)
- assignment end (`ended_at` set)
- assignment room change (tenant moves to another room, possibly another flat)

Algorithm (per affected tenant `T` and flat `F`):

1. Compute `should_be_member = has_active_assignment_in_flat(T, F)`.
2. If `should_be_member` and no active `ChatRoomMember`, insert member row (`joined_at = now`, `source = derived`).
3. If `!should_be_member` and active member row exists, set `left_at = now` (or delete member row if hard-remove policy; MVP prefers `left_at` for audit).

Flat membership rows use `role = member`; no tenant-editable role changes.

#### Validation

- Manual add/remove API for flat rooms returns `authorization_denied` for tenant principals.

---

### 3.2 Direct Messaging

**Requirements:** `FR-CM-003`, `FR-CM-005`, `FR-CM-008`

#### Open-or-create

1. Caller supplies `other_tenant_id`.
2. Resolve canonical pair `(low_id, high_id)`.
3. `SELECT` direct room by pair; if absent, create `ChatRoom{ kind: direct }` plus two `ChatRoomMember` rows in one transaction.

#### Behavior

- Conversation list shows other participant’s chat profile nickname/avatar.
- Unread count uses read cursor (§3.5).

---

### 3.3 Group Chats

**Requirements:** `FR-CM-004`

#### Create

- Creator becomes `ChatRoomMember` with `role = owner`.
- Initial members optional; creator can add tenants post-create.

#### Manage membership

| Action | Who | Effect |
|--------|-----|--------|
| Add member | owner | insert active member |
| Remove member | owner | set `left_at` on target |
| Leave | self | set `left_at` on self |
| Post | active member | allowed |

Removed or departed members: cannot post; historical messages remain.

#### Metadata

- `name`, `avatar_url` (or storage key) on `ChatRoom` for `kind = group`.
- Owner may update metadata.

---

### 3.4 Chat Profile

**Requirements:** `FR-CM-006`

#### Fields

| Field | Rules |
|-------|--------|
| `nickname` | optional override; display falls back to tenant legal name |
| `bio` | text, max length per validation policy |
| `avatar` | stored blob reference; MIME allowlist; max bytes |

#### Propagation

- Reads join `ChatTenantProfile` at render time for headers, member lists, and message author chips.
- No message row rewrite on profile change.

---

### 3.5 Messages, History, and Read State

**Requirements:** `FR-CM-007`, `FR-CM-008`, `FR-CM-009`

#### Send contract

```
POST /api/chat/rooms/{roomId}/messages
{
  "body": "string",
  "client_message_id": "uuid"   // required for idempotency
}
```

Processing:

1. Verify active membership.
2. If `client_message_id` already exists for (`room_id`, `author_tenant_id`), return existing message (200/201 per policy).
3. Insert `ChatMessage`; return payload with server `id` and `created_at`.
4. Fan-out to subscribers (§3.6).

#### History query

- Cursor pagination on `(created_at, id)`.
- Default tenant timeline: `ORDER BY created_at ASC, id ASC`.
- Page size cap (for example 50).

#### Read cursor

On conversation open:

- `last_read_message_id = MAX(visible message id in room)` (or latest at open time).
- Unread count = messages with `id > last_read_message_id` excluding self-authored messages (configurable; default excludes own messages).

---

### 3.6 Real-Time Delivery

**Requirements:** `FR-CM-009`

#### Transport (implementation choice)

- **MVP acceptable:** short-interval polling on open room + conversation list refresh.
- **Preferred:** SSE or WebSocket channel scoped to tenant session, emitting `message.created` and `room.read_updated` events.

#### Client retry

- Failed send keeps optimistic UI row in `failed` state.
- Retry reuses same `client_message_id` until server acknowledges persistence.

#### Read indicators

- Minimum: list badge from read cursor.
- Optional: `ChatMessageReceipt` rows per (`message_id`, `tenant_id`) — defer unless product requires per-message ticks in v1.

---

### 3.7 HTML Surface and APIs

**Requirements:** `FR-CM-010`, `FR-CM-011`

#### Routes (v1 minimum)

| Route | Purpose |
|-------|---------|
| `GET /chat/view` | conversation workspace |
| `GET /chat/view/rooms/{id}` | room timeline + send form |
| `POST /chat/view/rooms/{id}/messages` | send message (form) |
| `GET /chat/view/profile` | profile edit form |
| `POST /chat/view/profile` | save nickname/bio/avatar |
| `GET /chat/view/groups/new` | create group form (optional dedicated page) |
| `POST /chat/view/groups` | create group |
| `POST /chat/view/rooms/{id}/members` | add/remove group members |

#### Parity rule

HTML handlers and `/api/chat/*` handlers call the same application services (`ChatService` or equivalent); templates do not embed business rules.

---

### 3.8 Security and Audit

**Requirements:** `FR-CM-012`, `FR-CM-013`

#### Authorization matrix (baseline)

| Action | Tenant member | Non-member | Staff |
|--------|---------------|------------|-------|
| Read room messages | yes | no | policy |
| Post message | yes (active) | no | no |
| Edit flat membership | no | no | diagnostics only |
| Manage group members | owner | no | policy |
| Edit own chat profile | yes | — | — |

#### Audit (minimum)

- `chat.profile.update`
- `chat.group.member.remove` (owner action)
- `chat.moderation.*` (if staff moderation added)

---

## 4. Interface-Level Contracts

### 4.1 List APIs

Conversation list:

- sort by `last_message_at DESC` (denormalized on room or computed)
- include `unread_count`, `title`, `avatar_url`, `kind`, `last_message_preview`

Message list:

- `before` / `after` cursor on `(created_at, id)`

### 4.2 Mutation APIs

Send message, open DM, create group, membership changes, profile update:

- single transactional unit per request where multiple rows change
- return updated aggregates when useful (unread count, member list)

### 4.3 Error Semantics

- `validation_error`
- `authorization_denied`
- `not_found` (room hidden from non-members)
- `not_a_member`
- `room_closed` (if archived)
- `duplicate_client_message` (maps to success with existing row per idempotency policy)
- `avatar_invalid_type` / `avatar_too_large`

---

## 5. Non-Functional Specification

Mapped to `NFR-CM-001` through `NFR-CM-005`.

### 5.1 Data Integrity

- partial unique indexes for flat and direct rooms
- membership check in same transaction as message insert

### 5.2 Performance

- index `ChatMessage(room_id, created_at DESC)`
- index `ChatRoomMember(tenant_id)` partial `WHERE left_at IS NULL` for conversation list

### 5.3 Security

- room IDs are UUIDs; access always verified via membership, not obscurity alone

---

## 6. Cross-Module Interaction Specification

### 6.1 Administration Integration

**Event sources** (application hooks or domain events):

- `RoomAssignmentCreated`
- `RoomAssignmentEnded`
- `RoomAssignmentRoomChanged`

**Handler:** `SyncFlatChatMembership` resolves flat(s) for affected tenant(s) and applies §3.1 algorithm.

Administration does not write `ChatRoom` or `ChatRoomMember` directly.

### 6.2 Platform Integration

- Tenant session resolves `tenant_id` for all chat operations.
- Avatar bytes stored via platform file/blob helper.
- Optional notification events: `chat.message.created` for offline tenants — consumed by platform notification workers.

### 6.3 Forum

- No data coupling; cross-linking (open DM from forum profile) is UX-only via tenant id route.

---

## 7. Compliance Mapping

| Specification section | Requirement mapping |
| --------------------- | ------------------- |
| Flat provisioning and membership | FR-CM-001, FR-CM-002 |
| Direct messaging | FR-CM-003, FR-CM-005, FR-CM-008 |
| Group chats | FR-CM-004 |
| Chat profile | FR-CM-006 |
| Messages and read state | FR-CM-007, FR-CM-008, FR-CM-009 |
| HTML and APIs | FR-CM-010, FR-CM-011 |
| Security and audit | FR-CM-012, FR-CM-013 |
| Non-functional | NFR-CM-001 – NFR-CM-005 |

---

## 8. Implementation Notes

- Prefer explicit `ChatRoom.kind` discriminator over parallel per-type tables.
- Store canonical direct pair columns (`tenant_low_id`, `tenant_high_id`) with CHECK `tenant_low_id < tenant_high_id` for enforceable uniqueness.
- Denormalize `ChatRoom.last_message_at` and optional `last_message_preview` on insert for fast conversation list (update in same transaction as message create).
- Flat membership sync should be idempotent and safe to replay on assignment backfill jobs.
- Use `client_message_id` uniqueness scoped to (`room_id`, `author_tenant_id`).
- Align snake_case column names in SQL/GORM migrations with documentation names.

---

## 9. Cross-References

- Use cases: [chat.md](../use-cases/chat.md) (UC-CM-01–04)
- User stories: [chat.md](../user-stories/chat.md) (CM-001–005)
- Features: [chat-features.md](../features/chat-features.md)
- Database: [chat.md](../database/chat.md)
- Administration assignments: [administration-office.md](administration-office.md) §6.2, [administration.md](../database/administration.md) (`Flat`, `RoomAssignment`)
