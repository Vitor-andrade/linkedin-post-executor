<div align="center">

# LinkedIn Post Executor

**Crie, agende e publique conteúdo no seu LinkedIn — local, gratuito e sob seu controle.**

Uma ferramenta _open source_ e _local-first_ que ajuda desenvolvedores a manter uma presença
profissional consistente no LinkedIn sem gastar horas escrevendo nem um centavo em serviços.

[Por que existe](#-por-que-este-projeto-existe) ·
[Como funciona](#-como-funciona) ·
[Princípios](#-princípios-inegociáveis) ·
[Stack](#-stack-técnica) ·
[Arquitetura](#-arquitetura) ·
[Roadmap](#-roadmap)

![status](https://img.shields.io/badge/status-em%20construção-yellow)
![license](https://img.shields.io/badge/license-MIT-blue)
![local-first](https://img.shields.io/badge/local--first-✓-green)
![cost](https://img.shields.io/badge/custo-R$%200-brightgreen)

</div>

---

## 💡 Por que este projeto existe

Todo desenvolvedor sabe que estar presente no LinkedIn ajuda — abre portas, atrai
oportunidades e fortalece a reputação técnica. O problema não é a vontade de postar;
é o **tempo e a fricção**:

- Pensar no tema, escrever, formatar, achar as hashtags certas, lembrar de postar no horário certo.
- A formatação nativa do LinkedIn é limitada — negrito e itálico só com truques de Unicode.
- Manter **consistência** semana após semana é o que realmente move o ponteiro, e é exatamente
  onde a maioria desiste.

O **LinkedIn Post Executor** nasceu para remover essa fricção: você descreve a ideia,
a ferramenta gera um post bem estruturado e pronto para engajar, e — se você quiser —
publica ou agenda diretamente no **seu** perfil. Sem planilhas, sem copiar e colar manual,
sem assinaturas mensais.

> A intenção é simples: **transformar "eu deveria postar" em "já está agendado" em menos de um minuto.**

---

## ✨ O que ele faz

- 📝 **Geração de conteúdo assistida** — informe um título e uma breve descrição; a ferramenta
  produz um post completo, com gancho ("hook"), corpo escaneável, CTA e hashtags relevantes.
- 🎨 **Formatação nativa do LinkedIn** — aplica negrito/itálico via tipografia Unicode
  (𝗯𝗼𝗹𝗱, 𝘪𝘵𝘢𝘭𝘪𝘤), separadores e listas que renderizam direto no editor, sem Markdown.
- ✍️ **Edição manual sempre no controle** — todo texto gerado é um rascunho editável.
  A IA é um ponto de partida, não a palavra final.
- 📅 **Agendamento** — escolha data e hora; a ferramenta publica por você no momento certo.
- 🚀 **Publicação direta** — integração oficial via OAuth do LinkedIn, no seu próprio perfil.
- 🔌 **IA do seu jeito** — funciona _de graça_ com um modelo local (Ollama) ou, se preferir
  mais qualidade, conecte sua própria chave de API (Claude, OpenAI).

---

## ⚙️ Como funciona

```
   Você                 LinkedIn Post Executor (na sua máquina)              LinkedIn
   ────                 ───────────────────────────────────────              ────────
    │  título +            │                                                     │
    │  descrição           │                                                     │
    │ ───────────────────▶ │  gera com IA (Ollama local ou sua chave)            │
    │                      │ ─────────────┐                                      │
    │   post formatado     │ ◀────────────┘                                      │
    │ ◀─────────────────── │                                                     │
    │  revisa / edita      │                                                     │
    │ ───────────────────▶ │  salva rascunho (SQLite)                            │
    │                      │                                                     │
    │  "publicar agora"    │                                                     │
    │   ou "agendar"       │                                                     │
    │ ───────────────────▶ │  publica via OAuth ──────────────────────────────▶ │  ✅ post no ar
    │                      │  (ou no horário agendado)                           │
```

Tudo roda na sua máquina. Suas ideias, rascunhos e credenciais **nunca saem do seu computador** —
exceto a chamada que publica o post no LinkedIn, feita diretamente do seu app para a API oficial.

---

## 🔒 Princípios inegociáveis

Estas não são metas; são **regras** que guiam cada decisão técnica do projeto:

| Princípio | O que significa na prática |
|---|---|
| **100% open source** | Código aberto sob licença MIT. Sem caixas-pretas. |
| **Local-first** | Roda inteiramente na sua máquina. Você abre `localhost`, usa, e pronto. |
| **Custo zero** | Funciona sem pagar nada: IA local grátis + banco em arquivo + sem servidores. |
| **Traga suas credenciais (BYO)** | Você registra seu próprio app no LinkedIn e autoriza seu próprio perfil. A ferramenta **nunca** posta em nome de terceiros. |
| **Privacidade por padrão** | Nenhuma telemetria, nenhum intermediário. Seus dados são seus. |

Uma consequência elegante do modelo **BYO + local-first**: como cada pessoa autoriza o próprio
perfil com o próprio app, o projeto **não depende** de aprovações de "partner" da API do LinkedIn
(exigidas por ferramentas que postam por terceiros). Cada instância é soberana.

---

## 🧱 Stack técnica

A stack foi escolhida priorizando **distribuição simples, execução local e custo zero** —
não modismo. As decisões completas, com trade-offs, estão registradas nos
[ADRs](docs/architecture/).

| Camada | Tecnologia | Por quê |
|---|---|---|
| **Backend** | **Go** | Compila em **um único binário** sem dependências de runtime — baixe e rode. Concorrência (goroutines) ideal para o agendador que roda em segundo plano. |
| **Frontend** | **React + Vite** | UI moderna, embutida no binário Go via `go:embed`. Um processo serve tudo. |
| **Persistência** | **SQLite** | Banco em um arquivo local. Zero infraestrutura, zero custo, dados portáteis. |
| **IA** | **Ollama (padrão) + BYO key** | Camada plugável: modelo local gratuito por padrão; conecte Claude/OpenAI se quiser. |
| **Integração** | **LinkedIn OAuth** (`w_member_social`) | Publicação oficial no seu próprio perfil. |

> 💡 **Distribuição:** o resultado final é um executável único (ou um `docker run`). Sem instalar
> Node, sem subir banco, sem orquestrar serviços. Esse é o coração da proposta "local-first".

---

## 🏛 Arquitetura

```
╔══════════════════════════════════════════════════════════╗
║      Um binário Go  (ou  docker run)  —  http://localhost   ║
╠══════════════════════════════════════════════════════════╣
║                                                            ║
║   ┌─ React + Vite (UI)  ── embutida via go:embed            ║
║   │        ↕ HTTP local (JSON)                              ║
║   ├─ HTTP API                                               ║
║   │     ├── /api/generate   → geração de conteúdo (IA)      ║
║   │     ├── /api/drafts     → rascunhos e versões           ║
║   │     ├── /api/schedule   → fila de agendamento           ║
║   │     ├── /api/linkedin   → OAuth + publicação            ║
║   │     └── /api/settings   → chaves e preferências         ║
║   │                                                         ║
║   ├─ Scheduler (goroutine)  → dispara posts no horário      ║
║   │                                                         ║
║   ├─ AI Provider (interface plugável)                       ║
║   │     ├── Ollama  (local, grátis — padrão)                ║
║   │     └── API     (Claude / OpenAI — sua chave)           ║
║   │                                                         ║
║   └─ SQLite  → drafts · scheduled_posts · oauth_tokens      ║
║                                                            ║
╚══════════════════════════════════════════════════════════╝
```

Para o documento completo de arquitetura, com diagramas interativos (contexto do sistema,
componentes, fluxo de dados, deployment) e registros de decisão, veja
**[`docs/`](docs/)**.

---

## 🗺 Roadmap

O projeto evolui em fatias verticais — cada uma entrega valor de ponta a ponta.

- [x] **Fundação** — scaffold Go + Vite + SQLite, CI/CD, base do projeto.
- [ ] **Fatia 1 — Gerar** — informar tema → IA (Ollama) gera → revisar/editar → salvar rascunho. _(API e UI base prontas; falta persistir o rascunho gerado)_
- [ ] **Fatia 2 — Publicar** — OAuth do LinkedIn + publicação imediata no próprio perfil.
- [ ] **Fatia 3 — Agendar** — agendamento com o scheduler em segundo plano.
- [ ] **Refino** — suporte a chave própria (Claude/OpenAI), métricas locais, melhorias de UX.
- [ ] **Release** — binários multiplataforma e imagem Docker.

---

## 🚀 Rodando localmente

### Pré-requisitos

| Ferramenta | Para quê | Obrigatório? |
|---|---|---|
| [Go](https://go.dev/) 1.25+ | Compilar o binário | ✅ |
| [Node.js](https://nodejs.org/) 22+ e npm | Compilar a UI | ✅ |
| [Ollama](https://ollama.com/) | IA local gratuita (geração de conteúdo) | Opcional¹ |
| App no [LinkedIn Developers](https://developer.linkedin.com/) | Publicar no seu perfil | Opcional² |

> ¹ Necessário apenas para gerar conteúdo com IA local. Alternativamente, traga sua própria
> chave (Claude/OpenAI) — _em desenvolvimento_.
> ² Necessário apenas para a publicação automática — _em desenvolvimento_.

### Opção A — binário único (produção)

Compila a UI, embute no binário Go e executa tudo em um único processo:

```bash
make build      # gera a UI e compila o binário "linkedin-post-executor"
make run        # sobe em http://localhost:8080
```

Depois é só abrir **http://localhost:8080** no navegador.

### Opção B — modo desenvolvimento (hot reload)

Sobe a API Go e a UI do Vite em processos separados (a UI faz _proxy_ de `/api` para a API):

```bash
# terminal 1 — API Go (porta 8080)
make dev-api

# terminal 2 — UI Vite com hot reload (porta 5173)
make dev-ui
```

Acesse **http://localhost:5173**.

### IA local com Ollama (opcional)

Para a geração de conteúdo funcionar com o provedor padrão (custo zero):

```bash
# instale o Ollama (https://ollama.com) e baixe um modelo
ollama pull llama3.1
```

O app conversa com o Ollama em `http://localhost:11434` por padrão.

### Configuração

Todas as variáveis têm valor padrão; copie `.env.example` para `.env` apenas se quiser ajustar:

```bash
cp .env.example .env
```

| Variável | Padrão | Descrição |
|---|---|---|
| `LPE_ADDR` | `:8080` | Endereço HTTP local |
| `LPE_DB` | `data.db` | Caminho do arquivo SQLite |
| `LPE_AI_PROVIDER` | `ollama` | Provedor de IA |
| `LPE_OLLAMA_URL` | `http://localhost:11434` | URL do Ollama |
| `LPE_OLLAMA_MODEL` | `llama3.1` | Modelo do Ollama |

### Comandos úteis

```bash
make help     # lista todos os comandos disponíveis
make test     # roda os testes Go
make lint     # go vet + checagem de tipos da UI
make clean    # remove binário e banco local
```

---

## 🤝 Contribuindo

Contribuições são bem-vindas. Como o projeto preza por simplicidade e pelos princípios acima,
abra uma _issue_ para discutir mudanças significativas antes de enviar um _pull request_.

## 📄 Licença

Distribuído sob a licença **MIT**. Veja [`LICENSE`](LICENSE) para mais detalhes.
