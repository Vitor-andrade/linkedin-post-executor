# ADR-002 — Persistência com SQLite (arquivo local)

- **Status:** Aceito
- **Data:** 2026-05-25

## Contexto

A ferramenta roda localmente e deve ter **custo zero** e **zero infraestrutura**. Ainda assim,
é necessária persistência: rascunhos, versões, posts agendados (que precisam sobreviver a
reinícios do app) e tokens de OAuth.

## Decisão

Usar **SQLite** como banco em arquivo local (ex.: `./data.db`). Driver recomendado:
**`modernc.org/sqlite`** (implementação em Go puro, sem CGO), para manter o cross-compile e o
binário único triviais.

## Alternativas consideradas

- **Sem banco (JSON/arquivos):** simples, mas frágil para consultas, versionamento e
  agendamento confiável.
- **Postgres/MySQL:** poderoso, mas exige um servidor de banco — viola "custo zero" e
  "local-first / zero infra".

## Consequências

**Positivas**
- Zero custo e zero infraestrutura; dados portáteis em um único arquivo.
- Banco relacional de verdade, com transações e consultas.
- `modernc.org/sqlite` evita CGO → mantém o binário único e o cross-compile.

**Negativas / trade-offs**
- Concorrência de escrita limitada (irrelevante para um app single-user local).
- Tokens de OAuth devem ser **criptografados** em repouso no banco.
