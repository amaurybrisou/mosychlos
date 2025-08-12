---
applyTo: '**/*.go'
---

# Endpoint Patterns

Follow existing patterns exactly. Never create a new pattern without approval.

## Search Endpoints

- Proto: Place `SearchXXX` after `ListXXX` in proto definition.
- HTTP: `GET /{plural}`.
- Tests: Mirror success/minimal/error patterns from existing endpoints.

## List Endpoints

- Proto: `rpc ListXXX(ListXXXRequest) returns (service.ListXXXResponse)`.
- HTTP: `GET /{parent}/{id}/{plural}`.

## Get Endpoints

- Proto: `rpc GetXXX(GetXXXRequest) returns (service.XXX)`.
- HTTP: `GET /{parent}/{id}/{plural}/{id}`.

## Upsert Endpoints

- Proto: `rpc UpsertXXX(UpsertXXXRequest) returns (service.XXX)`.
- HTTP: `PUT /{parent}/{id}/{plural}/{id}`.
- Always return complete object.

## Delete Endpoints

- Proto: `rpc DeleteXXX(DeleteXXXRequest) returns (google.protobuf.Empty)`.
- HTTP: `DELETE /{parent}/{id}/{plural}/{id}`.

**Pattern Process:**

1. Find similar endpoint in repo.
2. Copy proto definition, HTTP mapping, implementation.
3. Adapt only specific names/fields.
4. Maintain error handling and test structure.
