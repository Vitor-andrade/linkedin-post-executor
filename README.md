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

- [ ] **Fundação** — scaffold Go + Vite + SQLite, CI/CD, base do projeto.
- [ ] **Fatia 1 — Gerar** — informar tema → IA (Ollama) gera → revisar/editar → salvar rascunho.
- [ ] **Fatia 2 — Publicar** — OAuth do LinkedIn + publicação imediata no próprio perfil.
- [ ] **Fatia 3 — Agendar** — agendamento com o scheduler em segundo plano.
- [ ] **Refino** — suporte a chave própria (Claude/OpenAI), métricas locais, melhorias de UX.
- [ ] **Release** — binários multiplataforma e imagem Docker.

---

## 🚀 Começando

> ⚠️ O projeto está em construção. As instruções de instalação e execução serão preenchidas
> conforme as primeiras fatias forem entregues.

Pré-requisitos previstos para rodar localmente:

- [Go](https://go.dev/) (build do binário)
- [Ollama](https://ollama.com/) instalado localmente para a IA gratuita _(opcional se usar chave própria)_
- Um app registrado no [LinkedIn Developers](https://developer.linkedin.com/) com o produto
  _"Share on LinkedIn"_ habilitado _(necessário apenas para publicar)_

---

## 🤝 Contribuindo

Contribuições são bem-vindas. Como o projeto preza por simplicidade e pelos princípios acima,
abra uma _issue_ para discutir mudanças significativas antes de enviar um _pull request_.

## 📄 Licença

Distribuído sob a licença **MIT**. Veja [`LICENSE`](LICENSE) para mais detalhes.
