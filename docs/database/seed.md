# Dorman Database Seed Plan

The default seed loads a **fabricate SQL export** (`internal/seed/data/data.sql`) that populates realistic dorm data (~400 tenants, forum, chat, doorman, etc.). Development login accounts are added after the SQL load.

## Running the seed

Requires PostgreSQL (e.g. `docker compose up -d postgres`) and migrated schema.

```bash
# Default: embedded SQL export + dev accounts
go run ./cmd/seed

# Custom SQL file (e.g. your own fabricate export)
go run ./cmd/seed --sql /path/to/data.sql

# Older programmatic Go seeder (docs row targets below)
go run ./cmd/seed --legacy --reset
```

Environment:

```bash
DATABASE_URL='postgres://dev:dev@localhost:5432/dormatory_manager?sslmode=disable' \
  go run ./cmd/seed
```

## Development logins

After seeding, these accounts use password **`password`** (plaintext login scaffold):

| Email | Roles |
| ----- | ----- |
| `admin@dorm.local` | administrator, doorman, office_worker, director |
| `staff1@dorm.local` | administrator |
| `staff2@dorm.local` | director |
| `staff3@dorm.local` | office_worker |
| `staff4@dorm.local` | doorman |

The SQL export uses fabricate role names (`admin`, `manager`, …); the seeder renames them to application names (`administrator`, `director`, `office_worker`, `doorman`) before attaching dev accounts.

## SQL export notes

- The script sets `session_replication_role = replica` so inserts can run without strict FK order.
- It **truncates all application tables** at the start; each run replaces data.
- Bundled file: `internal/seed/data/data.sql` (from fabricate, group 9 export).

## Legacy Go seeder (optional)

`--legacy --reset` runs the older `internal/seed` Go implementation with RNG `--seed` (default `42`). See row targets below for intended volumes.

### Sizing anchors (legacy)


| Anchor                    | Value                 | Rationale                                          |
| ------------------------- | --------------------- | -------------------------------------------------- |
| Buildings                 | 3                     | Small-to-mid dorm complex                          |
| Flats per building        | ~30                   | Mix of 1–6 floors                                  |
| Rooms per flat            | ~3 (avg)              | Mostly 2–4 bedrooms                                |
| Tenants per occupied room | ~1.3                  | Mostly singles, some doubles                       |
| Active period             | ~12 months of history | Drives time-series volumes                         |
| Staff users               | ~5                    | Admins, managers, reception, maintenance, security |


Implementation: `[internal/seed/](../../internal/seed/)`, entrypoint `[cmd/seed/main.go](../../cmd/seed/main.go)`.
