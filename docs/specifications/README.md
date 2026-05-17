# Module Specifications

Implementable behavior, invariants, interface contracts, and **cross-module integration** for each product area. Requirement IDs in each spec refer to the matching [requirements](../requirements/) document.

## Traceability

| Document | Requirements | Database | Features |
|----------|--------------|----------|----------|
| [administration-office.md](administration-office.md) | [administration-office.md](../requirements/administration-office.md) | [administration.md](../database/administration.md) | [administration-office-features.md](../features/administration-office-features.md) |
| [doorman.md](doorman.md) | [doorman.md](../requirements/doorman.md) | [doorman.md](../database/doorman.md) | [doorman-features.md](../features/doorman-features.md) |
| [forum.md](forum.md) | [forum.md](../requirements/forum.md) | [forum.md](../database/forum.md) | [forum-features.md](../features/forum-features.md) |
| [chat.md](chat.md) | [chat.md](../requirements/chat.md) | [chat.md](../database/chat.md) | [chat-features.md](../features/chat-features.md) |
| [platform.md](platform.md) | [platform.md](../requirements/platform.md) | [platform.md](../database/platform.md) | [platform-features.md](../features/platform-features.md) |

Use cases and stories: [../use-cases/](../use-cases/), [../user-stories/](../user-stories/).

## Cross-module sections

Every module specification includes **§6 Cross-Module Interaction Specification**. [platform.md](platform.md) §6 is the integration hub for shared auth, audit, and notification contracts; read it together with the module spec you are implementing.

## Documents

- `administration-office.md`
- `doorman.md`
- `forum.md`
- `chat.md`
- `platform.md`
