# ADR-004 — Publicação BYO via OAuth local-first (sem multi-tenant)

- **Status:** Aceito
- **Data:** 2026-05-25

## Contexto

Ferramentas que publicam no LinkedIn **em nome de terceiros** dependem da Community Management
API / Marketing Developer Platform, que exige aprovação rigorosa de "partner". Isso seria o
maior risco do projeto se ele fosse um SaaS multi-tenant.

O produto, porém, é local-first e individualizado: cada usuário roda a própria instância.

## Decisão

Adotar o modelo **BYO (bring your own credentials)**: cada usuário registra o **próprio app**
no LinkedIn Developers (produto _"Share on LinkedIn"_) e autoriza o **próprio perfil** via OAuth
com o escopo **`w_member_social`**. A ferramenta publica exclusivamente no perfil do dono da
instância — **nunca** em nome de terceiros.

Os tokens (access + refresh) são guardados **criptografados** no SQLite local, com **refresh
automático** (tokens do LinkedIn expiram em ~60 dias) para não quebrar agendamentos longos.

## Consequências

**Positivas**
- Elimina a necessidade de aprovação de "partner" do LinkedIn — o maior risco do projeto.
- Cada instância é soberana; credenciais nunca passam por um intermediário.
- Alinhado com privacidade e local-first.

**Negativas / trade-offs**
- Onboarding exige que o usuário registre um app no LinkedIn Developers (documentar bem).
- Gestão de expiração/refresh de token é responsabilidade da aplicação.
