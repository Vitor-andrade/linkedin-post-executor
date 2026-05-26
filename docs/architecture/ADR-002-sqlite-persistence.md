# ADR-002 — Persistence with SQLite (local file)

- **Status:** Accepted
- **Date:** 2026-05-25

## Context

The tool runs locally and must have **zero cost** and **zero infrastructure**. Even so,
persistence is needed: drafts, versions, scheduled posts (which must survive app
restarts), and OAuth tokens.

## Decision

Use **SQLite** as a local file-based database (e.g., `./data.db`). Recommended driver:
**`modernc.org/sqlite`** (a pure-Go implementation, no CGO), to keep cross-compilation and the
single binary trivial.

## Alternatives considered

- **No database (JSON/files):** simple, but fragile for queries, versioning, and reliable
  scheduling.
- **Postgres/MySQL:** powerful, but requires a database server — this violates "zero cost" and
  "local-first / zero infra".

## Consequences

**Positive**
- Zero cost and zero infrastructure; data is portable in a single file.
- A real relational database, with transactions and queries.
- `modernc.org/sqlite` avoids CGO → keeps the single binary and cross-compilation intact.

**Negative / trade-offs**
- Limited write concurrency (irrelevant for a single-user local app).
- OAuth tokens must be **encrypted** at rest in the database.
