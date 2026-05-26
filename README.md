<div align="center">

# LinkedIn Post Executor

**Create, schedule and publish content to your LinkedIn — local, free and under your control.**

An _open source_, _local-first_ tool that helps developers keep a consistent professional
presence on LinkedIn without spending hours writing or a single cent on services.

[Why it exists](#-why-this-project-exists) ·
[How it works](#-how-it-works) ·
[Principles](#-non-negotiable-principles) ·
[Stack](#-tech-stack) ·
[Architecture](#-architecture) ·
[Roadmap](#-roadmap)

![status](https://img.shields.io/badge/status-in%20progress-yellow)
![license](https://img.shields.io/badge/license-MIT-blue)
![local-first](https://img.shields.io/badge/local--first-✓-green)
![cost](https://img.shields.io/badge/cost-%240-brightgreen)

</div>

---

## 💡 Why this project exists

Every developer knows that being present on LinkedIn helps — it opens doors, attracts
opportunities and strengthens your technical reputation. The problem isn't the willingness to
post; it's the **time and friction**:

- Coming up with the topic, writing, formatting, finding the right hashtags, remembering to post
  at the right time.
- LinkedIn's native formatting is limited — bold and italic only work through Unicode tricks.
- Staying **consistent** week after week is what actually moves the needle, and it's exactly
  where most people give up.

**LinkedIn Post Executor** was born to remove that friction: you describe the idea, the tool
generates a well-structured, engagement-ready post, and — if you want — publishes or schedules it
directly to **your** profile. No spreadsheets, no manual copy-paste, no monthly subscriptions.

> The goal is simple: **turn "I should post" into "it's already scheduled" in under a minute.**

---

## ✨ What it does

- 📝 **AI-assisted content generation** — give it a title and a short description; the tool
  produces a complete post, with a hook, a scannable body, a CTA and relevant hashtags.
- 🎨 **Native LinkedIn formatting** — applies bold/italic via Unicode typography
  (𝗯𝗼𝗹𝗱, 𝘪𝘵𝘢𝘭𝘪𝘤), separators and lists that render directly in the editor, without Markdown.
- ✍️ **Manual editing, always in control** — every generated text is an editable draft.
  AI is a starting point, not the final word.
- 📅 **Scheduling** — pick a date and time; the tool publishes for you at the right moment.
- 🚀 **Direct publishing** — official integration via LinkedIn OAuth, to your own profile.
- 🔌 **AI your way** — works _for free_ with a local model (Ollama) or, if you prefer higher
  quality, connect your own API key (Claude, OpenAI).

---

## ⚙️ How it works

```
   You                  LinkedIn Post Executor (on your machine)              LinkedIn
   ───                  ────────────────────────────────────────             ────────
    │  title +             │                                                     │
    │  description         │                                                     │
    │ ───────────────────▶ │  generates with AI (local Ollama or your key)       │
    │                      │ ─────────────┐                                      │
    │   formatted post     │ ◀────────────┘                                      │
    │ ◀─────────────────── │                                                     │
    │  review / edit       │                                                     │
    │ ───────────────────▶ │  saves draft (SQLite)                               │
    │                      │                                                     │
    │  "publish now"       │                                                     │
    │   or "schedule"      │                                                     │
    │ ───────────────────▶ │  publishes via OAuth ────────────────────────────▶ │  ✅ post live
    │                      │  (or at the scheduled time)                         │
```

Everything runs on your machine. Your ideas, drafts and credentials **never leave your
computer** — except the call that publishes the post to LinkedIn, made directly from your app to
the official API.

---

## 🔒 Non-negotiable principles

These aren't goals; they are **rules** that guide every technical decision in the project:

| Principle | What it means in practice |
|---|---|
| **100% open source** | Open code under the MIT license. No black boxes. |
| **Local-first** | Runs entirely on your machine. You open `localhost`, use it, done. |
| **Zero cost** | Works without paying anything: free local AI + file-based database + no servers. |
| **Bring your own credentials (BYO)** | You register your own LinkedIn app and authorize your own profile. The tool **never** posts on behalf of third parties. |
| **Privacy by default** | No telemetry, no middleman. Your data is yours. |

An elegant consequence of the **BYO + local-first** model: since each person authorizes their own
profile with their own app, the project **does not depend** on LinkedIn "partner" approvals
(required by tools that post on behalf of third parties). Every instance is sovereign.

---

## 🧱 Tech stack

The stack was chosen by prioritizing **simple distribution, local execution and zero cost** — not
trends. The full decisions, with trade-offs, are recorded in the
[ADRs](docs/architecture/).

| Layer | Technology | Why |
|---|---|---|
| **Backend** | **Go** | Compiles into a **single binary** with no runtime dependencies — download and run. Concurrency (goroutines) is ideal for the scheduler that runs in the background. |
| **Frontend** | **React + Vite** | Modern UI, embedded into the Go binary via `go:embed`. One process serves everything. |
| **Persistence** | **SQLite** | A database in a single local file. Zero infrastructure, zero cost, portable data. |
| **AI** | **Ollama (default) + BYO key** | Pluggable layer: free local model by default; connect Claude/OpenAI if you want. |
| **Integration** | **LinkedIn OAuth** (`w_member_social`) | Official publishing to your own profile. |

> 💡 **Distribution:** the end result is a single executable (or a `docker run`). No installing
> Node, no spinning up a database, no orchestrating services. That's the heart of the
> "local-first" proposition.

---

## 🏛 Architecture

```
╔══════════════════════════════════════════════════════════╗
║      A single Go binary  (or  docker run)  —  http://localhost ║
╠══════════════════════════════════════════════════════════╣
║                                                            ║
║   ┌─ React + Vite (UI)  ── embedded via go:embed            ║
║   │        ↕ local HTTP (JSON)                              ║
║   ├─ HTTP API                                               ║
║   │     ├── /api/generate   → content generation (AI)      ║
║   │     ├── /api/drafts     → drafts and versions          ║
║   │     ├── /api/schedule   → scheduling queue             ║
║   │     ├── /api/linkedin   → OAuth + publishing           ║
║   │     └── /api/settings   → keys and preferences         ║
║   │                                                         ║
║   ├─ Scheduler (goroutine)  → fires posts at the right time ║
║   │                                                         ║
║   ├─ AI Provider (pluggable interface)                      ║
║   │     ├── Ollama  (local, free — default)                 ║
║   │     └── API     (Claude / OpenAI — your key)            ║
║   │                                                         ║
║   └─ SQLite  → drafts · scheduled_posts · oauth_tokens      ║
║                                                            ║
╚══════════════════════════════════════════════════════════╝
```

For the full architecture document, with interactive diagrams (system context, components, data
flow, deployment) and decision records, see **[`docs/`](docs/)**.

---

## 🗺 Roadmap

The project evolves in vertical slices — each one delivers end-to-end value.

- [x] **Foundation** — Go + Vite + SQLite scaffold, CI/CD, project base.
- [ ] **Slice 1 — Generate** — enter a topic → AI (Ollama) generates → review/edit → save draft. _(API and base UI ready; persisting the generated draft still pending)_
- [ ] **Slice 2 — Publish** — LinkedIn OAuth + immediate publishing to your own profile.
- [ ] **Slice 3 — Schedule** — scheduling with the background scheduler.
- [ ] **Polish** — bring-your-own-key support (Claude/OpenAI), local metrics, UX improvements.
- [ ] **Release** — cross-platform binaries and a Docker image.

---

## 🚀 Running locally

### Prerequisites

| Tool | What for | Required? |
|---|---|---|
| [Go](https://go.dev/) 1.25+ | Build the binary | ✅ |
| [Node.js](https://nodejs.org/) 22+ and npm | Build the UI | ✅ |
| [Ollama](https://ollama.com/) | Free local AI (content generation) | Optional¹ |
| App on [LinkedIn Developers](https://developer.linkedin.com/) | Publish to your profile | Optional² |

> ¹ Only needed to generate content with local AI. Alternatively, bring your own key
> (Claude/OpenAI) — _in progress_.
> ² Only needed for automatic publishing — _in progress_.

### Option A — single binary (production)

Builds the UI, embeds it into the Go binary and runs everything in a single process:

```bash
make build      # builds the UI and compiles the "linkedin-post-executor" binary
make run        # starts at http://localhost:8080
```

Then just open **http://localhost:8080** in your browser.

### Option B — development mode (hot reload)

Runs the Go API and the Vite UI as separate processes (the UI _proxies_ `/api` to the API):

```bash
# terminal 1 — Go API (port 8080)
make dev-api

# terminal 2 — Vite UI with hot reload (port 5173)
make dev-ui
```

Open **http://localhost:5173**.

### Local AI with Ollama (optional)

For content generation to work with the default provider (zero cost):

```bash
# install Ollama (https://ollama.com) and pull a model
ollama pull llama3.1
```

The app talks to Ollama at `http://localhost:11434` by default.

### Configuration

Every variable has a default; copy `.env.example` to `.env` only if you want to tweak something:

```bash
cp .env.example .env
```

| Variable | Default | Description |
|---|---|---|
| `LPE_ADDR` | `:8080` | Local HTTP address |
| `LPE_DB` | `data.db` | Path to the SQLite file |
| `LPE_AI_PROVIDER` | `ollama` | AI provider |
| `LPE_OLLAMA_URL` | `http://localhost:11434` | Ollama URL |
| `LPE_OLLAMA_MODEL` | `llama3.1` | Ollama model |

### Useful commands

```bash
make help     # list all available commands
make test     # run Go tests
make lint     # go vet + UI type checking
make clean    # remove the binary and the local database
```

---

## 🤝 Contributing

Contributions are welcome. Since the project values simplicity and the principles above, please
open an _issue_ to discuss significant changes before submitting a _pull request_.

## 📄 License

Distributed under the **MIT** license. See [`LICENSE`](LICENSE) for details.
