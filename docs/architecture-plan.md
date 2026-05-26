# LinkedIn Post Executor — Plano de Arquitetura

## Sumário executivo

O **LinkedIn Post Executor** é uma ferramenta **open source, local-first e de custo zero** que
ajuda um desenvolvedor a gerar, agendar e publicar conteúdo no **seu próprio** perfil do LinkedIn.
A arquitetura adotada é um **monólito modular** entregue como um **binário único em Go** com a
interface React/Vite embutida, persistência em **SQLite**, uma **camada de IA plugável**
(Ollama local por padrão, com opção de chave própria Claude/OpenAI) e publicação via **OAuth do
LinkedIn** no modelo _bring your own credentials_ (BYO).

> Veja os diagramas interativos em [`architecture-diagrams.html`](./architecture-diagrams.html)
> e as decisões em [`architecture/`](./architecture/).

## Resumo da descoberta (requisitos e restrições)

| Tema | Definição |
|---|---|
| **Problema** | Manter presença consistente no LinkedIn consome tempo; formatação e agendamento são fricções. |
| **Usuário** | O próprio desenvolvedor (single-user por instância). |
| **Modelo de produto** | Individualizado / self-hosted. "Vender" = cada pessoa roda a própria instância. |
| **Publicação** | Apenas no próprio perfil, via OAuth (`w_member_social`). Nunca para terceiros. |
| **Regras inegociáveis** | Open source (MIT) · local-first · custo zero · BYO credenciais · privacidade. |
| **Prioridades** | Qualidade e aprendizado acima de velocidade de entrega. Sem prazo. |

## Estilo de arquitetura

**Monólito modular** empacotado em um binário único.

| Estilo | Adequação | Veredito |
|---|---|---|
| **Monólito modular** | Single-user, local, simples; fronteiras de domínio claras (generate, drafts, schedule, linkedin). | ✅ **Escolhido** |
| Microserviços | Operação distribuída desnecessária para app local single-user. | ❌ Overkill |
| Serverless | Conflita com local-first e com o agendador persistente; depende de nuvem. | ❌ Não se aplica |

## Stack de tecnologia

| Camada | Primário | Alternativa | Racional |
|---|---|---|---|
| Backend | **Go** | Next.js full-stack | Binário único, distribuição trivial, goroutines para o scheduler. [ADR-001](./architecture/ADR-001-go-single-binary.md) |
| Frontend | **React + Vite** (embutido via `go:embed`) | HTMX/templ | UI rica, um único processo serve tudo. |
| Persistência | **SQLite** (`modernc.org/sqlite`) | arquivos JSON | Banco real, zero infra, binário sem CGO. [ADR-002](./architecture/ADR-002-sqlite-persistence.md) |
| IA | **Ollama (padrão) + BYO key** | só BYO key | Custo zero por padrão, flexível. [ADR-003](./architecture/ADR-003-pluggable-ai-ollama-default.md) |
| Publicação | **LinkedIn OAuth** (`w_member_social`) | — | BYO, sem aprovação de partner. [ADR-004](./architecture/ADR-004-byo-oauth-local-first.md) |
| Router HTTP | `chi` ou `net/http` (Go 1.22+) | — | Leve e idiomático. |
| CI | GitHub Actions (lint + test + build) | — | Qualidade e releases multiplataforma. |

## Arquitetura do sistema

### Contexto do sistema

```mermaid
graph TD
    user([Desenvolvedor<br/>dono da instância])
    subgraph local["Máquina local do usuário"]
        app["LinkedIn Post Executor<br/>(binário único Go)"]
        ollama["Ollama<br/>(LLM local — padrão)"]
        db[("SQLite<br/>data.db")]
    end
    llmapi["API de LLM externa<br/>(Claude / OpenAI — opcional, BYO key)"]
    linkedin["API do LinkedIn<br/>(Share on LinkedIn)"]

    user -->|"abre localhost,<br/>escreve, agenda"| app
    app -->|"gera conteúdo (padrão)"| ollama
    app -.->|"gera conteúdo (opcional, BYO key)"| llmapi
    app -->|"lê/grava rascunhos,<br/>agendamentos, tokens"| db
    app -->|"publica no próprio perfil<br/>OAuth w_member_social"| linkedin
```

### Componentes

```mermaid
graph TD
    subgraph binary["Binário único (Go) — http://localhost"]
        ui["UI React + Vite<br/>(embutida via go:embed)"]
        subgraph api["HTTP API"]
            gen["/api/generate"]
            drafts["/api/drafts"]
            sched["/api/schedule"]
            li["/api/linkedin"]
            settings["/api/settings"]
        end
        scheduler["Scheduler<br/>(goroutine + ticker)"]
        subgraph ai["AIProvider (interface)"]
            ollama["OllamaProvider<br/>(local, padrão)"]
            apiprov["APIProvider<br/>(Claude/OpenAI, BYO)"]
        end
        liclient["LinkedIn Client<br/>(OAuth + publish)"]
        repo["Repositório (SQLite)"]
    end
    dbfile[("data.db")]

    ui -->|JSON| api
    gen --> ai
    drafts --> repo
    sched --> repo
    settings --> repo
    li --> liclient
    liclient --> repo
    scheduler -->|"posts vencidos"| repo
    scheduler -->|publica| liclient
    repo --> dbfile
```

### Fluxo de dados

```mermaid
flowchart LR
    A["Usuário informa<br/>título + descrição"] --> B["/api/generate"]
    B --> C{Provedor<br/>de IA}
    C -->|padrão| D["Ollama (local)"]
    C -->|BYO key| E["Claude / OpenAI"]
    D --> F["Post formatado<br/>(Unicode, hook, CTA, hashtags)"]
    E --> F
    F --> G["Usuário revisa/edita"]
    G --> H["Salva rascunho<br/>(SQLite)"]
    H --> I{Ação}
    I -->|Publicar agora| J["LinkedIn Client → API"]
    I -->|Agendar| K["scheduled_posts<br/>(SQLite)"]
    K --> L["Scheduler (goroutine)<br/>verifica horário"]
    L -->|chegou a hora| J
    J --> M(["Post publicado<br/>no LinkedIn ✅"])
```

### Deployment

```mermaid
graph TD
    subgraph machine["Máquina do usuário (qualquer SO)"]
        subgraph proc["Processo único: linkedin-post-executor"]
            httpd["Servidor HTTP :PORT"]
            uiassets["Assets da UI (embed)"]
            sch["Scheduler (goroutine)"]
        end
        file[("data.db (arquivo SQLite)")]
        ollamasrv["Ollama (serviço local opcional)"]
        browser["Navegador → http://localhost:PORT"]
    end
    cloudllm["LLM em nuvem (opcional, BYO key)"]
    li["API do LinkedIn"]

    browser --> httpd
    httpd --> uiassets
    proc --> file
    proc --> ollamasrv
    proc -.->|opcional| cloudllm
    proc -->|HTTPS| li
```

### Sequência de publicação (agendada)

```mermaid
sequenceDiagram
    actor U as Usuário
    participant UI as UI (React)
    participant API as API (Go)
    participant DB as SQLite
    participant SCH as Scheduler
    participant LI as API LinkedIn

    Note over U,LI: Publicação agendada
    U->>UI: Define post + data/hora
    UI->>API: POST /api/schedule
    API->>DB: grava scheduled_post (status=pending)
    API-->>UI: 201 Created

    loop a cada tick
        SCH->>DB: busca posts vencidos (pending)
        DB-->>SCH: post(s) due
        SCH->>DB: lê + descriptografa token OAuth
        alt token expirado
            SCH->>LI: refresh token
            LI-->>SCH: novo access token
            SCH->>DB: grava token atualizado
        end
        SCH->>LI: POST /ugcPosts (w_member_social)
        LI-->>SCH: 201 + URN do post
        SCH->>DB: status=published, salva URN
    end
```

### Modelo de dados

```mermaid
erDiagram
    DRAFTS ||--o{ DRAFT_VERSIONS : "tem"
    DRAFTS ||--o| SCHEDULED_POSTS : "pode gerar"

    DRAFTS {
        int id PK
        string title
        string source_description
        string content
        string status
        datetime created_at
        datetime updated_at
    }
    DRAFT_VERSIONS {
        int id PK
        int draft_id FK
        string content
        string generated_by
        datetime created_at
    }
    SCHEDULED_POSTS {
        int id PK
        int draft_id FK
        string content
        datetime scheduled_for
        string status
        string linkedin_urn
        string error
        datetime created_at
    }
    OAUTH_TOKENS {
        int id PK
        string provider
        blob access_token_enc
        blob refresh_token_enc
        datetime expires_at
        datetime updated_at
    }
    SETTINGS {
        string key PK
        string value
    }
```

## Roadmap de evolução

Como o app é single-user local, não há "escala de usuários" tradicional. A evolução é por
**capacidade**, em fatias verticais:

1. **Fundação** — scaffold Go + Vite + SQLite, CI/CD, estrutura modular.
2. **Fatia 1 — Gerar** — tema → IA (Ollama) → revisar/editar → salvar rascunho.
3. **Fatia 2 — Publicar** — OAuth do LinkedIn + publicação imediata.
4. **Fatia 3 — Agendar** — scheduler em segundo plano.
5. **Refino** — BYO key (Claude/OpenAI), métricas locais, melhorias de UX.
6. **Release** — binários multiplataforma + imagem Docker.

## Análise de custo

| Componente | Custo |
|---|---|
| Compute | R$ 0 — roda na máquina do usuário |
| Banco de dados | R$ 0 — SQLite em arquivo |
| IA (padrão) | R$ 0 — Ollama local |
| IA (opcional) | Pay-per-use da chave do próprio usuário (BYO) |
| Hospedagem | R$ 0 — não há servidor |

O custo permanece **zero** por design. O único gasto possível é, opcionalmente, o consumo da
chave de LLM que o próprio usuário decidir conectar.

## Boas práticas e padrões

- **Fronteiras de domínio** explícitas por módulo (generate / drafts / schedule / linkedin).
- **Interface `AIProvider`** isolando o domínio de provedores externos (princípio de inversão de dependência).
- **TypeScript strict** no frontend; lint + format (Go: `golangci-lint`, `gofmt`).
- **Pirâmide de testes:** unit (formatação Unicode, agendamento), integração (repositório SQLite), e2e do fluxo gerar→publicar. Foco na lógica de negócio, sem dogma de cobertura.
- **ADRs** para decisões relevantes (este diretório).
- **Conventional Commits**, branches curtas, PRs revisados.

### Anti-padrões a evitar

- Otimização prematura (ex.: escolher tech por "performance" de CPU quando a latência é do LLM).
- Acoplar a lógica de geração a um provedor de IA específico.
- Guardar segredos/tokens em texto puro.

## Arquitetura de segurança

- **Segredos** (chaves de IA, client secret do LinkedIn) nunca no código — via `.env`/settings.
- **Tokens OAuth criptografados em repouso** no SQLite; refresh automático antes da expiração (~60 dias).
- Superfície mínima: API exposta apenas em `localhost`.
- Sem telemetria nem terceiros intermediando dados.

## Riscos e mitigações

| Risco | Impacto | Mitigação |
|---|---|---|
| Mudanças/limites da API do LinkedIn | Alto | Isolar em `LinkedIn Client`; tratar erros e rate limits; documentar versão da API. |
| Expiração de token quebrar agendamentos | Médio | Refresh automático + reautenticação guiada na UI. |
| Qualidade variável do Ollama local | Médio | Permitir BYO key; oferecer prompts/modelos recomendados. |
| Onboarding do app no LinkedIn ser confuso | Médio | Documentação passo a passo com prints. |
| Perda do arquivo `data.db` | Baixo | Orientar backup; arquivo único é fácil de copiar. |

## Decisões registradas (ADRs)

- [ADR-001 — Backend em Go, binário único](./architecture/ADR-001-go-single-binary.md)
- [ADR-002 — Persistência com SQLite](./architecture/ADR-002-sqlite-persistence.md)
- [ADR-003 — IA plugável, Ollama por padrão](./architecture/ADR-003-pluggable-ai-ollama-default.md)
- [ADR-004 — Publicação BYO via OAuth local-first](./architecture/ADR-004-byo-oauth-local-first.md)

## Próximos passos

1. Scaffold do repositório (estrutura Go modular + Vite + SQLite + CI + licença MIT).
2. Implementar a **Fatia 1 (Gerar)** com o `OllamaProvider` e o prompt do `linkedin-post-writer`.
3. Implementar OAuth + publicação (Fatia 2) e, em seguida, o scheduler (Fatia 3).
