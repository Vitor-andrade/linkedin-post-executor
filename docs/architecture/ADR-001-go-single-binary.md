# ADR-001 — Go backend compiled into a single binary

- **Status:** Accepted
- **Date:** 2026-05-25

## Context

The project is **local-first**, **open source**, and must be distributed as an
**individualized product** (each user runs their own instance). The ease of
"download and run" is a first-order requirement. There is also a scheduler that needs
to run in the background to publish posts at the scheduled time.

The author is proficient in Node/TypeScript and is interested in going deeper into Go, with no time pressure.

## Decision

Use **Go** for the backend, compiling into a **single binary** that embeds the UI (React/Vite)
via `go:embed`. The same process serves the interface, exposes the local HTTP API, and runs the
scheduler.

## Alternatives considered

- **Next.js full-stack:** faster and more familiar to deliver, but local execution requires the
  Node runtime, and a persistent local scheduler becomes awkward (it would need a separate
  `node-cron` worker). Distribution turns into "clone the repo + npm install" or Docker.
- **Go API + separate Next.js:** more flexible, but it means two processes to run locally,
  which weakens the "download and run" proposition.

## Consequences

**Positive**
- Trivial distribution: a single executable with no runtime dependencies; simple cross-compilation.
- Goroutines make the background scheduler natural and lightweight.
- Low resource usage when running locally.
- Meets the author's learning goal.

**Negative / trade-offs**
- Learning curve and more boilerplate than Next.js.
- The frontend ecosystem remains JS either way (the embedded Vite build).

> ⚠️ **Important:** the choice of Go is **not** about CPU performance. The perceived latency is
> dominated by LLM generation and the LinkedIn API calls. The real motivation is the
> **distribution model** and the scheduler's **persistent process**.
