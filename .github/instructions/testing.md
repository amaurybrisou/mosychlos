---
applyTo: '**/*.go'
---

# Testing Standards

- **Table-driven tests**:

```go
cases := []struct{name string; ...}{...}
for _, c := range cases {
    t.Run(c.name, func(t *testing.T) {
        ...
    })
}
```

- Test naming: `TestStruct_Method` or `TestPackage_Function`.
- Table slice named `cases`, loop variable `c`.
- No hyphens in test case names.
- Mock all interfaces.
- Test cases: success, minimal, error.

**CRITICAL**

- Never test with main packages file => always \_test pattern

## Datastore Test Suites

- Use `suite.Run(t, new(DatastoreTestSuite))`.
- Each test is a method: `func (suite *DatastoreTestSuite) TestX()`.
- Containerized setup/teardown via testcontainers.
