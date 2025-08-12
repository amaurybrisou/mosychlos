---
applyTo: '**/*.go'
---

# Protocol Buffers

- Enums: uppercase snake case, `UNSPECIFIED` at 0.
- CRUD naming: Get, List, Upsert, Create, Update, Delete.
- Non-persistent logic: IsXxx, GenXxx.
- Requests/Responses: unique messages per RPC unless scoped.
- Upsert must return complete object.
- Document each RPC in gateway with `openapi`.
- Message ordering:
  - Models first.
  - Request/response second.
