---
applyTo: '**/*.go'
---

# Linting Rules

- Blank line before `return`, `if`, `range`, `defer` (unless after assignment).
- No cuddled variable declarations.
- Validate bounds before numeric conversions.
- Use protobuf getters (e.g., `GetId()`), never direct field access.
- Avoid unnecessary conversions when types already match.
- Max line length: 120 chars.
- Functions ≤ 60 lines; extract helpers.
- Pre-allocate slices when size is known.
