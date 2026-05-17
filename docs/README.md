# Documentation

Product and engineering documentation for **Dormitory Manager**, organized by artifact type and product module. Module boundaries match the root [README](../README.md#product-modules): Administration and Office, Doorman, Tenant Forum, Tenant Chat, plus shared **platform** concerns.

## Documentation layers

Read top-down for *what* to build; bottom-up from code changes for *where* to record behavior.

| Layer | Folder | Purpose |
|-------|--------|---------|
| Features | [features/](features/) | Product-facing capability catalog by module |
| Requirements | [requirements/](requirements/) | Functional and non-functional requirements (`FR-*`) |
| Use cases | [use-cases/](use-cases/) | Actor goals and flows (`UC-*`) |
| User stories | [user-stories/](user-stories/) | Backlog stories with acceptance criteria |
| Specifications | [specifications/](specifications/) | Implementable behavior, APIs, invariants, cross-module contracts |
| Database | [database/](database/) | Module-scoped tables, keys, and cross-module dependencies |

**Cross-cutting** (RBAC, audit, notifications, health) is documented under the **platform** module in each layer (`FR-CC-*`, `UC-CC-*`, `CC-*`, `cross-cutting.md` where applicable).

## Module index

| Module | Features | Requirements | Use cases | User stories | Specification | Database |
|--------|----------|--------------|-----------|--------------|---------------|----------|
| Administration and Office | [administration-office-features.md](features/administration-office-features.md) | [administration-office.md](requirements/administration-office.md) | [administration-office.md](use-cases/administration-office.md) | [administration-office.md](user-stories/administration-office.md) | [administration-office.md](specifications/administration-office.md) | [administration.md](database/administration.md) |
| Doorman | [doorman-features.md](features/doorman-features.md) | [doorman.md](requirements/doorman.md) | [doorman.md](use-cases/doorman.md) | [doorman.md](user-stories/doorman.md) | [doorman.md](specifications/doorman.md) | [doorman.md](database/doorman.md) |
| Forum | [forum-features.md](features/forum-features.md) | [forum.md](requirements/forum.md) | [forum.md](use-cases/forum.md) | [forum.md](user-stories/forum.md) | [forum.md](specifications/forum.md) | [forum.md](database/forum.md) |
| Chat | [chat-features.md](features/chat-features.md) | [chat.md](requirements/chat.md) | [chat.md](use-cases/chat.md) | [chat.md](user-stories/chat.md) | [chat.md](specifications/chat.md) | [chat.md](database/chat.md) |
| Platform | [platform-features.md](features/platform-features.md) | [platform.md](requirements/platform.md) | [cross-cutting.md](use-cases/cross-cutting.md) | [cross-cutting.md](user-stories/cross-cutting.md) | [platform.md](specifications/platform.md) | [platform.md](database/platform.md) |

Requirement ID prefixes: `FR-AO-*`, `FR-DM-*`, `FR-FM-*`, `FR-CM-*`, `FR-CC-*`.

## Other folders

| Folder | Status | Notes |
|--------|--------|-------|
| [agent/](agent/) | Active | Iteration digest and agent handoff notes for maintainers |
| [dev/](dev/) | Placeholder | Engineer onboarding (see [dev/README.md](dev/README.md)) |
| [user/](user/) | Placeholder | Staff and pilot user handbook |
| [testing/](testing/) | Placeholder | System test case catalog and testing docs |

Branch workflow and merge gates are described in the repository root README and (when added) `docs/dev/branch-workflow.md`.
