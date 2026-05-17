# Database documentation

This directory holds **module-scoped** database declarations: tables (or views) each module relies on, their **purpose**, and **relationships** to other tables. It complements the global entity list in [data-model.md](../specifications/data-model.md), which remains the canonical high-level inventory.

## Module documents


| Document                               | Module                    |
| -------------------------------------- | ------------------------- |
| [administration.md](administration.md) | Administration and Office |
| [doorman.md](doorman.md)               | Doorman operations        |
| [forum.md](forum.md)                   | Forum module              |
| [chat.md](chat.md)                     | Chat module               |
| [platform.md](platform.md)             | Platform / cross-cutting  |

Each module document includes a **Cross-Module Dependencies** (or boundaries) section describing how tables relate to other modules.

## Related specifications

- [Data model specification](../specifications/data-model.md) — core entities and cross-cutting rules (DM-01–DM-03)
- [Glossary](../specifications/glossary.md) — Flat, Room, Tenant, assignments, chat linkage
- [Architecture](../specifications/architecture.md) — PostgreSQL, GORM, migrations, data access style

