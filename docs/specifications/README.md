# Module Specifications

This folder defines implementable behavior, invariants, interface contracts, and **cross-module integration** for each product area.

## Traceability

| Document | Requirements | Database |
|----------|--------------|----------|
| [administration-office.md](administration-office.md) | [FR-01](../requirements/administration-office.md) | [administration.md](../database/administration.md) |
| [doorman.md](doorman.md) | [FR-02](../requirements/doorman.md) | [doorman.md](../database/doorman.md) |
| [forum.md](forum.md) | [FR-03](../requirements/forum.md) | [forum.md](../database/forum.md) |
| [chat.md](chat.md) | [FR-04](../requirements/chat.md) | [chat.md](../database/chat.md) |
| [platform.md](platform.md) | [FR-05](../requirements/platform.md) | [platform.md](../database/platform.md) |

## Cross-module sections

Every module specification includes **§6 Cross-Module Interaction Specification** (platform spec §6 is the integration hub). Pair with [platform.md](platform.md) for shared auth, audit, and notification contracts.
