---
applyTo: '**/*.go'
---

# Logging Standards

- Include essential context: IDs, parameters, counts, durations.
- Consistent field names:
  - `user_id`, `campaign_id`, `organization_id`
  - Arrays: plural (`campaign_ids`)
- Levels:
  - Error: unrecoverable failures.
  - Warn: recoverable/config issues.
  - Info: major events.
  - Debug: detailed execution flow.

## Performance Logging

- Measure operation duration.
- Warn if >5s for critical paths.

## Don't Log:

- Sensitive data (passwords, tokens).
- Large payloads (log size/count instead).
