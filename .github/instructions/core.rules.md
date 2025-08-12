---
applyTo: '**/*.go'
---

# Core Development Rules

- **Never modify code without explicit permission**: Only proceed when told “Fix it”, “Make the change” or similar.
- **Investigation vs Implementation**:
  - Investigation: Analyze code, identify issues, explain findings.
  - Implementation: Any code change, test update, or modification of constants/configs.
- If an issue is found:
  ```
  I found an issue: ...
  Analysis: ...
  Possible approaches: ...
  What would you like me to do?
  ```
- Always run tests with `-race`.
- Use `any` instead of `interface{}`.
- Add comments only when they add real value:
  - Function comments start **uppercase**.
  - Inline comments start **lowercase**.
- Follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).
- Branches: `feat/`, `fix/`, `chore/` — lowercase, no underscores.
