---
applyTo: '**/*.go'
---

# NATS Conventions

- Producer stream: service name lowercase.
- Subject: service.resource (snake_case).
- Consumer name: `<service>_<resource>`.
- Durable name = consumer name lowercase.
- Use `WithMsgID()` for deduplication.
- For inter-service calls: no `MaxDeliver`, monitor streams, fix issues.
