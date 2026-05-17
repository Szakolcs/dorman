# Dorman Database Seed Plan

A realistic dormitory management seed sized around ~3 buildings / ~400 tenants. Volumes are tuned so the system feels lived-in (a year or two of activity) without exploding child tables.

### Sizing anchors


| Anchor                    | Value                 | Rationale                                          |
| ------------------------- | --------------------- | -------------------------------------------------- |
| Buildings                 | 3                     | Small-to-mid dorm complex                          |
| Flats per building        | ~30                   | Mix of 1–6 floors                                  |
| Rooms per flat            | ~3 (avg)              | Mostly 2–4 bedrooms                                |
| Tenants per occupied room | ~1.3                  | Mostly singles, some doubles                       |
| Active period             | ~12 months of history | Drives time-series volumes                         |
| Staff users               | ~5                    | Admins, managers, reception, maintenance, security |


## Dependency-ordered groups & row targets

### Group 1 — Foundation (no FK deps)


| Table     | Rows | Notes                                                        |
| --------- | ---- | ------------------------------------------------------------ |
| buildings | 3    | 3 named dorm buildings with codes (e.g. A, B, C)             |
| roles     | 6    | admin, manager, reception, maintenance, security, tenant     |
| users     | 225  | 25 staff + 400 tenant-linked accounts; mix of principal_type |



| Table | Rows | Notes                                           |
| ----- | ---- | ----------------------------------------------- |
| flats | 50   | ~30 per building, distributed across floors 0–5 |
| rooms | 100  | Avg 3.3 rooms/flat, capacity 1–3, ~5% archived  |



| Table                | Rows | Notes                                                          |
| -------------------- | ---- | -------------------------------------------------------------- |
| tenants              | 100  | 80% linked to a user; degree (BSc/MSc/PhD), faculty, age 18–30 |
| user_roles           | 100  | One per staff/tenant + ~25 multi-role staff                    |
| chat_tenant_profiles | 100  | 1:1 with tenants (unique constraint)                           |
| room_assignments     | 320  | ~400 active + ~120 historical move-outs                        |



| Table      | Rows | Notes                                              |
| ---------- | ---- | -------------------------------------------------- |
| activities | 50   | Long-running programs (gym slot, study room, etc.) |
| events     | 200  | One-off events spread over 18 months               |



| Table                 | Rows | Notes                                                  |
| --------------------- | ---- | ------------------------------------------------------ |
| inventory_items       | 200  | ~600 in rooms/flats, ~200 building-level common assets |
| item_loans            | 100  | ~150 currently checked out, rest returned              |
| maintenance_tickets   | 200  | Mix of statuses; ~70% closed                           |
| ticket_status_changes | 150  | Avg 2.5 transitions per ticket                         |
| operational_jobs      | 100  | Cleaning/inspection/maintenance assignments            |
| packages              | 300  | ~85% picked up                                         |
| package_notifications | 100  | Avg 1.5 per package across email/sms/push              |



| Table               | Rows | Notes                                             |
| ------------------- | ---- | ------------------------------------------------- |
| tenant_entry_tokens | 500  | Avg 1.25 per tenant (some renewed)                |
| access_events       | 1000 | ~20 entries/exits per tenant; weighted by daytime |
| guest_visits        | 200  | ~2 per active tenant over the period              |
| guest_access_events | 200  | Check-in + check-out per visit                    |



| Table                     | Rows | Notes                                               |
| ------------------------- | ---- | --------------------------------------------------- |
| chat_rooms                | 290  | 90 flat rooms (1:1 with flat) + 400 direct DM pairs |
| chat_room_members         | 600  | Flat: ~4 members each; DMs: 2 each; some historical |
| chat_messages             | 2000 | Avg ~30 per room, skewed (some rooms much busier)   |
| chat_membership_sync_logs | 600  | One per move-in/move-out + room rebalances          |



| Table                    | Rows | Notes                                                        |
| ------------------------ | ---- | ------------------------------------------------------------ |
| forum_posts              | 800  | Mix of kind: announcement, event, activity, discussion, poll |
| forum_post_schedules     | 250  | Roughly 1:1 with event-kind posts                            |
| forum_post_updates       | 300  | A few updates on busier posts                                |
| forum_polls              | 100  | Subset of posts that are poll-kind                           |
| forum_poll_options       | 350  | Avg 3.5 options per poll                                     |
| forum_poll_votes         | 100  | ~25 votes per poll on average                                |
| forum_attendance_intents | 400  | Going/maybe/no for event posts                               |
| forum_comments           | 400  | Avg ~4 per post; some threaded replies                       |
| forum_reactions          | 200  | Likes etc. across posts/comments                             |
| forum_post_views         | 1000 | Many views per published post                                |
| forum_moderation_actions | 100  | Rare moderation events                                       |



| Table        | Rows | Notes                                             |
| ------------ | ---- | ------------------------------------------------- |
| audit_events | 1000 | Broad coverage of admin/system actions over 18 mo |


## Running the seed

Requires PostgreSQL (e.g. `docker compose up -d postgres`) and migrated schema.

```bash
# Default: postgres://dev:dev@localhost:5432/dormatory_manager?sslmode=disable
go run ./cmd/seed --reset

# Reproducible RNG + custom database URL
DATABASE_URL='postgres://dev:dev@localhost:5432/dormatory_manager?sslmode=disable' \
  go run ./cmd/seed --reset --seed 42
```

Flags:

- `--reset` — truncate all application tables before inserting (CASCADE).
- `--seed` — RNG seed (default `42`).

Implementation: `[internal/seed/](../../internal/seed/)`, entrypoint `[cmd/seed/main.go](../../cmd/seed/main.go)`.

All seeded users share password `password` (matches the current login scaffold).