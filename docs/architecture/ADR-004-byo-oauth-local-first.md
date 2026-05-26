# ADR-004 — BYO publishing via local-first OAuth (no multi-tenancy)

- **Status:** Accepted
- **Date:** 2026-05-25

## Context

Tools that publish to LinkedIn **on behalf of third parties** depend on the Community Management
API / Marketing Developer Platform, which requires strict "partner" approval. This would be the
project's biggest risk if it were a multi-tenant SaaS.

The product, however, is local-first and individualized: each user runs their own instance.

## Decision

Adopt the **BYO (bring your own credentials)** model: each user registers their **own app**
on LinkedIn Developers (the _"Share on LinkedIn"_ product) and authorizes their **own profile** via OAuth
with the **`w_member_social`** scope. The tool publishes exclusively to the instance owner's
profile — **never** on behalf of third parties.

The tokens (access + refresh) are stored **encrypted** in local SQLite, with **automatic
refresh** (LinkedIn tokens expire in ~60 days) so that long-running schedules don't break.

## Consequences

**Positive**
- Eliminates the need for LinkedIn "partner" approval — the project's biggest risk.
- Each instance is sovereign; credentials never pass through an intermediary.
- Aligned with privacy and local-first.

**Negative / trade-offs**
- Onboarding requires the user to register an app on LinkedIn Developers (document this well).
- Managing token expiration/refresh is the application's responsibility.
