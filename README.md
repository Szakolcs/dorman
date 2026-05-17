# Dormitory Manager

Dormitory Manager is a modular web platform for dormitory operations, staff workflows, and tenant community engagement.

## Tech Stack

- **Backend:** Go + Echo
- **SSR UI Components:** Templ
- **Frontend Interactivity:** HTMX
- **Styling:** Tailwind CSS
- **Database:** PostgreSQL
- **Testing:** Go's built-in `testing` package + `testify` for assertions and test suites

## Product Modules

### 1) Administration and Office Module

Used by administrators and office workers to manage:

- Rooms
- Inventory
- Maintenance tickets
- Tenants
- Job scheduling
- News and articles for tenants
- Activities (for example ping pong, basketball)

### 2) Doorman Module

Used by doormen to manage:

- Tenant packages
- Tenant notifications
- Guest access for tenants
- Dorm access via QR code reader
- Lending activity keys/accessories (for example gym key, basketballs)

### 3) Tenant Forum Module

Used by tenants and student consult to:

- Read dorm news
- Advertise and discover activities
- Organize events
- Vote in polls
- Like and comment on news/events
- Mark attendance intent for events

### 4) Tenant Chat Module

A lightweight Discord-like chat experience:

- One automatic room per flat
- Private messages
- Group chats (activities/friend groups)
- Profile customization:
  - Avatar
  - Nickname (full name by default)
  - Bio

## Branching Model

End-to-end flow (requirements on **`docs/*`**, optional **`project-setup/*`**, **`development/*`**, **`testing/unit|integration|system`**, then **`docs/dev`** / **`docs/user`** back to **`docs/main`**): **[`docs/dev/branch-workflow.md`](docs/dev/branch-workflow.md)**.

Base branches (each line is a branch name; work happens on these or on topic branches below):

- `docs/main`
- `project-setup/main`
- `development/main`
- `testing/main`
- `beta/main`
- `release/main`

### Feature Development Flow

1. Create an issue for the feature.
2. Branch from `development/main` using:
   - `#-feature/feature-name`
3. Implement the feature.
4. Open a PR and merge into `testing/main`.

### Testing Branch Flow (Per Feature)

Each testing stage has a dedicated issue and pull request:

1. `testing/unit/feature-name` (created after feature is in `testing/main`)
2. `testing/integration/feature-name` (created from unit branch)
3. `testing/system/feature-name` (created from integration branch)

After testing is complete:

- Merge to `development/main` HEAD
- Merge to `beta/main`

## Git and PR Rules

- **One feature per branch**
- **One issue + one PR per feature**
- **Dedicated issue + PR for each testing stage**
- Keep commits compact, logical, and readable (atomic commits are not required)

## Commit Convention

Use conventional commits in this format:

`$tag($scope):Brief description of what changed and why`

Examples:

- `feat(auth):add dorm staff login to secure admin module`
- `fix(tickets):prevent duplicate maintenance request creation`
- `test(chat):cover group room creation edge cases`

## Testing Strategy

Testing layers map to your branch flow:

- **Unit tests:** fast isolated tests for business logic
- **Integration tests:** repository/service/API boundary tests with PostgreSQL test fixtures
- **System tests:** end-to-end behavior validation across module workflows

Recommended tools:

- `go test` for test execution
- `testify` for expressive assertions and suite organization
- Optional DB test setup via containerized PostgreSQL
- **SQLite-backed tests (`testing/unit` branch):** some admin service/store suites use temporary SQLite via `gorm.io/driver/sqlite` (**CGO-enabled**; install GCC or Clang).

**Integration (`testing/integration`):** Postgres-backed seams live under `internal/integration/admintest/` with build tag **`integration`** (see [TQ-01](docs/specifications/testing-and-quality.md)). They **skip** unless `INTEGRATION_DATABASE_URL` or `DATABASE_URL` points at Postgres (migrate + truncate before each scenario). Run **`make integration-test`** or:

```bash
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/dormatory_manager?sslmode=disable
go test -tags=integration -count=1 ./internal/integration/admintest/...
```

Demo seed assertions can take ~1–3 minutes loading the full catalogue.

**System (`testing/system`):** full-router journeys live under **`internal/systemtest/admintest/`** with build tag **`system`**. Prefer **`SYSTEM_DATABASE_URL`** for CI separation; **`INTEGRATION_DATABASE_URL`** / **`DATABASE_URL`** are accepted (see [`internal/systemtest/admintest/setup.go`](internal/systemtest/admintest/setup.go)). Human-readable case IDs are in **[`docs/testing/system-test-cases.md`](docs/testing/system-test-cases.md)**.

- **`make system-test`** — runs system tests against Postgres (~3–30+ minutes depending on demo seed scenarios).
- **`make system-test-report`** — writes **`test-output/go-test-system-json.ndjson`** plus a Markdown summary **`test-output/system-test-report.md`** via **`cmd/gen-system-report`** (`test-output/` is gitignored).

- **`make full-test-report`** — runs **Unit** (`go test ./... -v -json`), **Integration** (`-tags=integration`), and **System** (`-tags=system`) into **`test-output/unit.ndjson`** (and counterparts), then aggregates **`test-output/full-test-report.md`** via **`cmd/gen-full-test-report`**. The generator exits **0** even when tests fail (artifact is still written); pass **`-strict`** to `go run ./cmd/gen-full-test-report …` in CI if the process should fail the job.

```bash
export SYSTEM_DATABASE_URL=postgres://postgres:postgres@localhost:5432/dormatory_manager?sslmode=disable
make system-test
# or Markdown report:
make system-test-report

# All tiers + combined Markdown (Postgres URLs may be empty—integration/system rows will SKIP):
make full-test-report
cat test-output/full-test-report.md
```

## Demo database catalog (`cmd/seed`)

When PostgreSQL matches `DATABASE_URL`, you can populate a repeatable demo matching a large dorm (~500 residents):

- **`go run ./cmd/seed`** — migrates the administration schema then inserts the demo if it has not already been inserted.
- **`SEED_DEMO_DATA`** — when set to a truthy value (`1`, `true`, `yes`, `on`), the HTTP server loads the same catalog automatically after migrations on startup (idempotent skip if already present).

**What gets inserted (all migrated administration tables get rows):**

- Single building **Central Dormitory (demo seeded)** — used only to detect reruns (`internal/platform/seeding.SeedBuildingName`).
- **5** staff **`users`** (`seed-admin-1@dorm.local` … **`seed-admin-5@dorm.local`**) with deterministic principals (**`AdminUserUUID(0)…AdminUserUUID(4)`** in [`internal/platform/seeding`](internal/platform/seeding/demo.go)).
- **120 flats** and matching **`rooms`** (60 large + 60 small layouts), **`500` `tenants`**, **`500` active `room_assignments`**, **`40`** empty beds remaining.
- **`inventory_items`** — room-linked and shared-area (games lounge / gymnasium) examples; plus a **`inventory.patch`** path so **`audit_events`** includes inventory updates.
- **`maintenance_tickets`** (open hallway, room plumbing with full **`ticket_status_changes`** chain through closed, HVAC in-progress) and **`operational_jobs`** (scheduled/finished/cancelled, some linked to rooms or tickets).
- **`forum_posts`** — published bulletin + draft drill notice + package-hours post.
- **`activities`** — six published events; ping pong links to the sports bulletin **`forum_posts`** row (via **`activities.forum_post_id`** — unique FK).
- Many **`audit_events`** from **`room`** creation and **`room_assignment.create`**, transitions on maintenance, **`maintenance_ticket.assign`**, **`maintenance_ticket.status_change`**, and two explicit seed-only audit checks.

If the demo building already exists, the seed exits without inserts. After changing the demo catalog locally, recreate the Postgres volume or drop/recreate **`public`** in dev before **`go run ./cmd/seed`** so the fuller dataset runs again.

**Testing with the stub:**

```bash
# First administrator (stable across machines — see Demo testing in seeding/demo_test.go)
export STAFF_UID="3d49b55a-0948-5b8f-b8aa-dce36f621f2b"
curl -sS -H "X-Staff-User-Id: $STAFF_UID" -H "X-Staff-Role: administrator" http://localhost:8080/admin/rooms?limit=3
```

In development, **`/admin/view/dev/set-staff?user_id=$STAFF_UID&role=administrator`** also works for HTML admin.

## Documentation pointers

- **Integration / agent handoff** (testing gates → docs merge, task-cycle digest): [`docs/agent/README.md`](docs/agent/README.md)
- **Staff and pilot users** (web admin, demo data, sign-in quirks, stubs): [`docs/user/README.md`](docs/user/README.md)
- **Developer onboarding** (architecture, Docker/Air, admin HTTP surfaces, migrations, seeding, tests): [`docs/dev/README.md`](docs/dev/README.md)
- Domain glossary (flat vs room, chat linkage): [docs/specifications/glossary.md](docs/specifications/glossary.md)
- MVP phasing and non-goals: [docs/requirements/mvp-scope.md](docs/requirements/mvp-scope.md)

## Initial Architecture Direction

- Modular monolith first (clear package boundaries per module)
- Shared auth/authorization and tenant identity model
- Domain-driven service/repository separation
- HTMX-first server-rendered UX to minimize frontend complexity
- Strong test coverage gates before `beta` and `release`

## Roadmap (High Level)

1. Project setup and core architecture skeleton
2. Auth + role model (admin, office, doorman, tenant, Student Consult Organizer as scoped tenant; see [docs/requirements/roles-and-permissions.md](docs/requirements/roles-and-permissions.md))
3. Module-by-module implementation
4. Progressive test hardening by branch stage
5. Beta stabilization
6. Production release preparation