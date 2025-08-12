---
applyTo: '**/*.go'
---

# Validation Rules

- **Don't duplicate proto validation** (`buf.validate` already covers required, length, format, enums).
- Validate in Go only for:
  - Business rules not expressible in proto.
  - External dependencies (e.g., referenced IDs exist).
  - Cross-field logic.

## Do Validate in Go:

- Ownership rules.
- External resource existence.
- Complex interdependent fields.

## Don't Validate in Go:

- Required fields.
- String lengths.
- Formats.
- Enums.
