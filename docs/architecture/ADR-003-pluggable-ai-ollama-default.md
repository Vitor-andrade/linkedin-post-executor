# ADR-003 — Pluggable AI layer with Ollama as the default

- **Status:** Accepted
- **Date:** 2026-05-25

## Context

Content generation is the heart of the product, but the project has **zero cost** as a rule.
Cloud LLM providers (Claude, OpenAI) require a paid key. At the same time, anyone who wants
higher quality should be able to use their own key.

## Decision

Define a decoupled **`AIProvider` interface**, with interchangeable implementations:

- **`OllamaProvider`** — the default. A model run locally via [Ollama](https://ollama.com/),
  at zero cost.
- **`APIProvider`** — "bring your own key" (BYO) for Claude/OpenAI, configurable.

The provider choice lives in configuration (settings in SQLite / `.env`). The generation prompt
follows the specification of the `agents/linkedin-post-writer.agent.md` agent and always produces
posts in English.

## Alternatives considered

- **BYO key only:** simpler and better text quality, but requires a paid key for the
  app to work at all — this violates "zero cost".
- **Ollama only:** 100% free and private, but quality depends on the user's hardware and
  removes the flexibility for those who prefer a state-of-the-art model.

## Consequences

**Positive**
- Honors "zero cost" and "open source" with the local default.
- Flexible: switching providers does not touch the domain logic.
- Maximum privacy in the local option (nothing leaves the machine).

**Negative / trade-offs**
- Maintaining more than one implementation and normalizing differences between providers.
- The quality of the default (Ollama) varies with the local model/hardware.
