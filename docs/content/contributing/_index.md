---
title: Contributing
weight: 7
---

# Contributing

## Development Setup

```bash
git clone https://github.com/wakaba/wakaba.git
cd wakaba
make build
make test
```

## Architecture Principles

1. **Dependencies point inward** — adapters depend on ports, not the reverse
2. **Domain purity** — no Fiber, GORM, SQLite, or MCP imports in domain
3. **Event-first** — significant operations publish domain events
4. **Library-first** — core logic usable without HTTP server
5. **Testable** — ports enable mock/fake implementations

## Adding a New Feature

1. Define domain types in `internal/domain/`
2. Define port interfaces in `internal/ports/`
3. Implement adapters in `internal/adapters/`
4. Create application service in `internal/application/`
5. Wire into `cmd/wakaba/main.go`
6. Add tests

## Code Style

- Standard Go conventions
- No comments unless asked
- Table-driven tests with subtests
- Error wrapping with `fmt.Errorf("context: %w", err)`

## Testing

```bash
make test          # Run all tests
make test-race     # With race detector
make cover         # Coverage report
```

Tests use in-memory SQLite for isolation. Each test gets its own database.
