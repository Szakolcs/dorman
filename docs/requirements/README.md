# Module Requirements

Functional and non-functional requirements per product module. Each document includes **scope boundaries** (what this module owns vs. delegates) and a **traceability matrix** linking to user stories, use cases, and features.

Platform requirements (`FR-CC-*`) are the canonical source for RBAC, audit, notifications, and health; module-specific security FRs reference them.

## Documents

| Document | Module | FR prefix | Features |
|----------|--------|-----------|----------|
| [administration-office.md](administration-office.md) | Administration and Office | `FR-AO-*` | [administration-office-features.md](../features/administration-office-features.md) |
| [doorman.md](doorman.md) | Doorman | `FR-DM-*` | [doorman-features.md](../features/doorman-features.md) |
| [forum.md](forum.md) | Forum | `FR-FM-*` | [forum-features.md](../features/forum-features.md) |
| [chat.md](chat.md) | Chat | `FR-CM-*` | [chat-features.md](../features/chat-features.md) |
| [platform.md](platform.md) | Platform / cross-cutting | `FR-CC-*` | [platform-features.md](../features/platform-features.md) |

## Related artifacts

| Layer | Folder |
|-------|--------|
| Use cases | [../use-cases/](../use-cases/) |
| User stories | [../user-stories/](../user-stories/) |
| Specifications | [../specifications/](../specifications/) |
| Database | [../database/](../database/) |
