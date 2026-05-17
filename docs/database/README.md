# Database documentation

Module-scoped relational design: tables (and views) each module owns or depends on, column purposes, keys, and **cross-module dependencies**. Aligns with the matching [specifications](../specifications/) and [requirements](../requirements/) documents.

## Module documents

| Document | Module | Specification |
|----------|--------|---------------|
| [administration.md](administration.md) | Administration and Office | [administration-office.md](../specifications/administration-office.md) |
| [doorman.md](doorman.md) | Doorman operations | [doorman.md](../specifications/doorman.md) |
| [forum.md](forum.md) | Tenant forum | [forum.md](../specifications/forum.md) |
| [chat.md](chat.md) | Tenant chat | [chat.md](../specifications/chat.md) |
| [platform.md](platform.md) | Platform / cross-cutting | [platform.md](../specifications/platform.md) |

Each module document includes scope (owned vs. read-only tables) and a **Cross-Module Dependencies** (or boundaries) section.

## Conventions

- PostgreSQL-oriented naming; migrations may use snake_case columns.
- Identity, RBAC, and `AuditEvent` are defined in [platform.md](platform.md); product modules reference them rather than redefining them.
- Housing and assignment tables in administration drive forum visibility and chat flat-room membership as described in those modules’ boundary sections.

## Related documentation

- [Specifications](../specifications/) — domain invariants and integration contracts
- [Requirements](../requirements/) — `FR-*` traceability to tables where noted
