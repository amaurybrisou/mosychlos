---
applyTo: '**/*.go'
---

# Datastore Conventions

- Single `Datastore` interface per file.
- Storage-agnostic: no direct DB/driver types in interface.
- MongoDB:
  - Index query fields.
  - Use snake_case for collection/field names.
  - Avoid `omitempty` for optional fields to ensure they're sent.
  - Bulk ops: Ordered, fail fast, handle errors individually.
