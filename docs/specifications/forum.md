# Forum Module Specification

## Document Purpose

This specification translates Forum module requirements into implementable system behavior, data constraints, workflow contracts, and integration boundaries. Requirement IDs refer to [forum requirements](../requirements/forum.md). Table-level detail is in [forum database](../database/forum.md).

---

## 1. Module Context

The Forum module is the tenant-facing communication and engagement layer:

- unified feed for official and community content
- activities and events with attendance intent and organizer updates
- polls with eligibility, vote limits, and result visibility modes
- reactions and comment threads
- staff moderation for community safety
- HTML-first workspace at `/forum/view` with API parity

### 1.1 Actor Model

- **Tenant**: read feed, comment, react, vote, set attendance intent
- **Student Consult Organizer**: create/manage official polls and events (v1 may delegate via staff-stub per UC-FM-04)
- **Staff (Administrator / Office Worker)**: official publishing via administration; moderation
- **System**: enforces invariants, aggregates counters, applies visibility rules, emits audit

### 1.2 Module Boundaries

- **Administration and Office**: source of truth for official news/activity/event lifecycle; forum stores or mirrors tenant-visible `ForumPost` representations
- **Platform**: `User` / tenant identity, RBAC, `AuditEvent`, optional notification delivery
- **Doorman / notifications**: optional feed entries; doorman owns operational records

Out of scope:

- housing assignment, maintenance, inventory administration
- chat messaging and room membership
- doorman package and guest desk workflows

---

## 2. Domain Model and Core Invariants

Entities are detailed in `docs/database/forum.md`. Conceptual core types:

- `ForumPost` (feed item with `kind` discriminator)
- `ForumPostSchedule` (or embedded schedule fields for timed posts)
- `ForumPoll`, `ForumPollOption`, `ForumPollVote`
- `ForumComment`
- `ForumReaction`
- `ForumAttendanceIntent`
- `ForumPostView` (optional analytics)
- `ForumModerationAction`
- platform `User`, `Tenant`, `AuditEvent`
- administration linkage: `Activity`, `Event` (optional `forum_post_id`)

### 2.1 Invariants (mandatory)

- **I-FM-01 Publication visibility**: only `published` (and not staff-hidden) posts appear in tenant default feed.
- **I-FM-02 Pin ordering**: when sort is pinned-first, `pinned_at` / `pin_priority` determines order among pinned items.
- **I-FM-03 Single attendance intent**: at most one current `ForumAttendanceIntent` row per (`post_id`, `tenant_id`).
- **I-FM-04 Vote uniqueness**: poll configuration defines uniqueness; default single-choice is one vote row per (`poll_id`, `tenant_id`).
- **I-FM-05 Reaction uniqueness**: at most one reaction per (`target_type`, `target_id`, `user_id`).
- **I-FM-06 Result visibility**: `post_close` polls must not expose option vote counts to tenants before `closes_at`.
- **I-FM-07 Author ownership**: tenants may mutate only their own community posts and comments unless moderated or staff-overridden.
- **I-FM-08 Official immutability**: posts with `source = administration` are not editable by tenants in forum APIs.

---

## 3. Capability Specifications

### 3.1 Unified Feed and Post Detail

**Requirements:** `FR-FM-001`, `FR-FM-002`

#### Behavior

- Feed query returns published posts across kinds with pagination.
- Sort modes:
  - `newest`: `published_at DESC`, tie-break `id DESC`
  - `pinned_first`: pinned group first (`pin_priority DESC`, `pinned_at DESC`), then `published_at DESC`
- Detail load hydrates type-specific extensions (schedule, poll, attendance summary).

#### Validation

- Unpublished or `hidden` posts return `not_found` for tenant principals unless staff moderator.

#### Optional analytics

- On detail open, insert `ForumPostView` when policy enabled; failure is logged, not blocking.

---

### 3.2 Community Activities and Official Events

**Requirements:** `FR-FM-003`, `FR-FM-004`, `FR-FM-005`, `FR-FM-006`

#### Post kinds

| `kind` | Typical author | Notes |
|--------|----------------|-------|
| `official_news` | staff | from administration publish |
| `community_activity` | tenant | FM-002 |
| `event` | staff / organizer | capacity, deadline |
| `announcement` | staff | short-lived official note |

#### Community activity lifecycle

States (minimum):

- `draft` (optional)
- `published`
- `archived`

Tenant author may edit `title`, `body`, schedule fields while `published`; archive transitions to `archived`.

#### Attendance intent

- Values: `going`, `not_going`
- Upsert per (`post_id`, `tenant_id`)
- Expose aggregates: `going_count`, `not_going_count` on detail and organizer queries

#### Organizer updates

- Modeled as `ForumPostUpdate` rows or child posts linked via `parent_post_id` / `update_of_post_id`
- Visible per product policy (default: all viewers of parent event)

#### Administration integration

- On administration publish of news/activity/event, forum upserts `ForumPost` with `source = administration` and foreign keys to `Activity` / `Event` when present
- Forum does not own staff draft workflow; listens to published state from administration

---

### 3.3 Polling

**Requirements:** `FR-FM-007`, `FR-FM-008`, `FR-FM-009`

#### Behavior

- Poll is a `ForumPost` with `kind = poll` and child `ForumPoll` record.
- Organizer creates options (≥ 2) and `closes_at`.
- Tenant submits vote; server validates eligibility and mode.

#### Configuration

| Field | Values |
|-------|--------|
| `choice_mode` | `single`, `multiple` |
| `results_visibility` | `live`, `post_close` |
| `eligibility` | `all_active_tenants` (baseline), extensible |

#### Vote contract

- **Single:** insert one `ForumPollVote` with selected `option_id`; reject second vote from same tenant.
- **Multiple:** insert one row per selected option; reject duplicate (`poll_id`, `tenant_id`, `option_id`).

#### Results API

- `live`: return option counts to eligible viewers immediately
- `post_close`: return counts only when `now >= closes_at`; before close, return status without counts (or zeroed stub per UX policy)

#### Export

- Organizer export returns option labels, vote counts, and anonymized voter counts per policy (no PII in export unless staff role)

---

### 3.4 Reactions and Comments

**Requirements:** `FR-FM-010`, `FR-FM-011`, `FR-FM-012`

#### Reactions

- Default reaction type: `like`
- Toggle: POST creates reaction; DELETE removes
- Targets: `post`, `comment`

#### Comments

- Create on post; optional `parent_comment_id` for threading
- Author edit updates `body` and `edited_at`
- Author delete sets `deleted_at` (soft delete); content hidden in tenant views
- Staff `hide` / `remove` via moderation table sets `moderation_state`

#### Moderation

Actions (minimum):

- `hide` — excluded from feed, visible to moderators
- `restore` — returns to published visibility
- `remove` — terminal for community content; retains audit row

Each action appends `ForumModerationAction` and may emit `AuditEvent`.

---

### 3.5 HTML Surface and APIs

**Requirements:** `FR-FM-013`, `FR-FM-014`

#### Routes (v1 minimum)

| Route | Purpose |
|-------|---------|
| `GET /forum/view` | feed workspace |
| `GET /forum/view/posts/{id}` | post detail |
| `POST /forum/view/posts/{id}/comments` | comment form |
| `POST /forum/view/posts/{id}/reactions` | like toggle |
| `POST /forum/view/posts/{id}/attendance` | intent form |
| `POST /forum/view/polls/{id}/votes` | poll vote form |

#### Parity rule

HTML handlers and JSON/API handlers invoke the same application services; no duplicated business rules in templates.

---

### 3.6 Security and Audit

**Requirements:** `FR-FM-015`, `FR-FM-016`

#### Authorization matrix (baseline)

| Action | Tenant | Organizer | Staff |
|--------|--------|-----------|-------|
| Read published feed | yes | yes | yes |
| Community post CRUD (own) | yes | — | — |
| Official poll/event create | — | yes | yes |
| Moderate community content | — | — | yes |
| Vote eligible poll | yes | yes* | — |

\*Organizers who are also tenants follow tenant vote rules.

#### Audit (minimum)

- `forum.moderation.hide`
- `forum.moderation.restore`
- `forum.moderation.remove`
- `forum.poll.create`
- `forum.poll.close_override` (staff only, if supported)

---

## 4. Interface-Level Contracts

### 4.1 List APIs

Feed, comments, poll results (when visible), moderation queue:

- pagination (`limit`, `cursor` or `offset`)
- filters: `kind`, `official_only`, `community_only`, date range
- stable default sort as defined in §3.1

### 4.2 Mutation APIs

Vote, intent, reaction, comment, community post publish:

- single transactional unit per request
- return updated aggregates where useful (counts, user’s current vote/intent)

### 4.3 Error Semantics

- `validation_error`
- `authorization_denied`
- `not_found`
- `poll_closed`
- `already_voted`
- `ineligible_voter`
- `results_not_visible`
- `edit_not_allowed`
- `concurrency_conflict`

---

## 5. Non-Functional Specification

Mapped to `NFR-FM-001` through `NFR-FM-005`.

### 5.1 Data Integrity

- enforce uniqueness for votes, reactions, attendance intents
- use transactions when updating counters derived from child rows (or compute from indexed aggregates)

### 5.2 Performance

- index `ForumPost(published_at)`, partial index on `pinned_at IS NOT NULL`, `ForumPollVote(poll_id)`, `ForumComment(post_id)`

### 5.3 Security

- tenant-scoped operations bind to authenticated principal
- poll result leakage prevented in service layer for `post_close` mode

---

## 6. Cross-Module Interaction Specification

### 6.1 Administration Integration

- Published official content creates or updates `ForumPost` rows consumed by forum feed queries.
- Administration remains authoritative for draft/publish/cancel/postpone of official activities and events.
- Forum stores `administration_activity_id` / `administration_event_id` optional FKs for drill-back.

### 6.2 Platform Integration

- All actors resolve through platform identity tables.
- Optional notification events: `event.updated`, `poll.closed` — consumed by platform notification workers.

### 6.3 Doorman / Notifications (optional)

- Package-ready or similar notifications may appear as feed items with `kind = system_notice` if product enables cross-module fan-in.

---

## 7. Compliance Mapping

| Specification section | Requirement mapping |
| --------------------- | ------------------- |
| Unified feed and detail | FR-FM-001, FR-FM-002 |
| Activities and events | FR-FM-003 – FR-FM-006 |
| Polling | FR-FM-007 – FR-FM-009 |
| Reactions and comments | FR-FM-010, FR-FM-011 |
| Moderation | FR-FM-012 |
| HTML and APIs | FR-FM-013, FR-FM-014 |
| Security and audit | FR-FM-015, FR-FM-016 |
| Non-functional | NFR-FM-001 – NFR-FM-005 |

---

## 8. Implementation Notes

- Prefer a single `ForumPost` table with `kind` discriminator over parallel per-type tables; use extension tables (`ForumPoll`, schedule columns) for type-specific fields.
- Compute reaction and attendance counts from indexed aggregates or maintained counters updated in the same transaction as child inserts.
- Keep poll visibility checks in the service layer, not only in templates, so API and HTML paths stay consistent.
- Soft-delete comments with `deleted_at`; do not hard-delete rows referenced by moderation audit.
- Align snake_case column names in SQL/GORM migrations with documentation names.

---

## 9. Cross-References

- Use cases: [forum.md](../use-cases/forum.md) (UC-FM-01–04)
- User stories: [forum.md](../user-stories/forum.md) (FM-001–005)
- Features: [forum-features.md](../features/forum-features.md)
- Database: [forum.md](../database/forum.md)
- Administration publishing: [administration-office.md](administration-office.md) §3.6
