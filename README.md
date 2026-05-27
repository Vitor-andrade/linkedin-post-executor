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

- 💡 **Post-idea suggestions** — a sidebar with topic ideas by category (frontend, backend,
  databases, devops, security, AI, career) plus an AI "generate more" for a specific technology;
  click an idea to seed the title.
- 📝 **AI-assisted content generation** — give it a title and a short description; the tool
  produces a complete post, with a hook, a scannable body, a CTA and relevant hashtags.
- 🎨 **Native LinkedIn formatting** — applies bold/italic via Unicode typography
  (𝗯𝗼𝗹𝗱, 𝘪𝘵𝘢𝘭𝘪𝘤), separators and lists that render directly in the editor, without Markdown.
- ✍️ **Manual editing, always in control** — every generated text is an editable draft.
  AI is a starting point, not the final word.
- 🗂 **Draft management** — save, list, reopen, edit and delete drafts, all stored locally.
- 📅 **Scheduling** — pick a date and time; a background worker publishes for you at the right
  moment, retrying transient failures with exponential backoff.
- 🚀 **Direct publishing** — official integration via LinkedIn OAuth, to your own profile.
- 🔌 **AI your way** — works _for free_ with a local model (Ollama) or a free Gemini key, or
  connect your own key (Claude, OpenAI) if you prefer.
- 📊 **Local metrics** — drafts, scheduled, published and last-publish counts, derived entirely
  from your own data (no telemetry).
- 🎨 **Modern UI** — a clean gradient interface with light/dark support and in-app confirmation
  dialogs for anything irreversible.

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
| **AI** | **Ollama (default) + BYO key** | Pluggable layer: free local model by default; connect Gemini (free tier), Claude or OpenAI if you want. |
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
║   │     └── BYO key (Gemini / Claude / OpenAI)              ║
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
- [x] **Slice 1 — Generate** — enter a topic → AI generates → review/edit → save, list and edit drafts.
- [x] **Slice 2 — Publish** — LinkedIn OAuth + immediate publishing to your own profile.
- [x] **Slice 3 — Schedule** — scheduling with the background scheduler (retries failed posts with exponential backoff).
- [x] **Polish** — bring-your-own-key support (Gemini, Claude, OpenAI) and local metrics.
- [ ] **Release** — cross-platform binaries and a Docker image. _(deferred — see [`docs/TODO.md`](docs/TODO.md); personal/local use for now)_

---

## 🚀 Running locally

### Quick start (step by step)

```bash
# 1. Clone and enter the project
git clone https://github.com/Vitor-andrade/linkedin-post-executor.git
cd linkedin-post-executor

# 2. (optional) free local AI — install Ollama from https://ollama.com, then:
ollama pull llama3.1

# 3. (optional) configure credentials — copy the template and edit as needed
cp .env.example .env

# 4. Build the UI + single binary and run it
make build
make run

# 5. Open the app
#    → http://localhost:8080
```

That's the whole install. Steps 2 and 3 are optional: without Ollama you can still write/edit
posts manually or use a cloud key (see [AI providers](#-ai-providers)); without LinkedIn
credentials everything works except publishing/scheduling.

To publish to LinkedIn, do the one-time setup in
[Publishing to LinkedIn](#-publishing-to-linkedin), then click **Connect LinkedIn** in the app.

### Prerequisites

| Tool | What for | Required? |
|---|---|---|
| [Go](https://go.dev/) 1.25+ | Build the binary | ✅ |
| [Node.js](https://nodejs.org/) 22+ and npm | Build the UI | ✅ |
| [Ollama](https://ollama.com/) | Free local AI (content generation) | Optional¹ |
| App on [LinkedIn Developers](https://developer.linkedin.com/) | Publish to your profile | Optional² |

> ¹ Only needed to generate content with the default local AI. Alternatively, bring your own key
> — a **free Gemini** key, or Claude/OpenAI. See [AI providers](#-ai-providers).
> ² Only needed for publishing and scheduling. See [Publishing to LinkedIn](#-publishing-to-linkedin).

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

Every variable has a default. Copy `.env.example` to `.env` and edit what you need — the file is
loaded automatically at startup (real environment variables take precedence over it):

```bash
cp .env.example .env
```

| Variable | Default | Description |
|---|---|---|
| `LPE_ADDR` | `:8080` | Local HTTP address |
| `LPE_DB` | `data.db` | Path to the SQLite file |
| `LPE_AI_PROVIDER` | `ollama` | AI provider: `ollama` \| `gemini` \| `anthropic` \| `openai` |
| `LPE_OLLAMA_URL` | `http://localhost:11434` | Ollama URL |
| `LPE_OLLAMA_MODEL` | `llama3.1` | Ollama model |
| `LPE_GEMINI_API_KEY` | — | Gemini key (when provider is `gemini`) |
| `LPE_GEMINI_MODEL` | `gemini-2.0-flash` | Gemini model |
| `LPE_ANTHROPIC_API_KEY` | — | Claude key (when provider is `anthropic`) |
| `LPE_OPENAI_API_KEY` | — | OpenAI key (when provider is `openai`) |
| `LPE_LINKEDIN_CLIENT_ID` | — | LinkedIn app client ID |
| `LPE_LINKEDIN_CLIENT_SECRET` | — | LinkedIn app client secret |
| `LPE_LINKEDIN_REDIRECT_URL` | `http://localhost:8080/api/linkedin/callback` | OAuth redirect (must match the app) |
| `LPE_KEY_FILE` | OS config dir | File holding the auto-generated token-encryption key |

See `.env.example` for the full list with comments.

---

## 🤖 AI providers

The content generator is pluggable. Pick one with `LPE_AI_PROVIDER`; all of them reuse the same
LinkedIn-tuned system prompt, so the formatting is identical regardless of the engine.

| Provider | Value | Key | Cost |
|---|---|---|---|
| **Ollama** (default) | `ollama` | — | Free, fully local |
| **Gemini** | `gemini` | `LPE_GEMINI_API_KEY` | **Free tier** |
| **Claude** | `anthropic` (alias `claude`) | `LPE_ANTHROPIC_API_KEY` | Pay-per-use |
| **OpenAI** | `openai` | `LPE_OPENAI_API_KEY` | Pay-per-use |

For a free cloud option, grab a Gemini key at
[aistudio.google.com/apikey](https://aistudio.google.com/apikey) and set:

```bash
LPE_AI_PROVIDER=gemini
LPE_GEMINI_API_KEY=your-key
```

The active provider is shown in the app header (e.g. `AI: gemini (gemini-2.0-flash)`).

---

## 🔗 Publishing to LinkedIn

Publishing uses **your own** LinkedIn app (bring your own credentials), so the project needs no
partner approval and never posts on behalf of anyone else.

> **Posts go to your personal profile**, not to a Company Page. The page below is required only to
> *create* the developer app (LinkedIn's rule) — a placeholder page is fine and nothing is posted
> to it. Posting to a Company Page would need different scopes and partner approval, which this
> project intentionally avoids.

1. Go to [LinkedIn Developers](https://www.linkedin.com/developers/apps) and **create an app**
   (it must be associated with a Company Page you manage — create a quick placeholder one if needed).
2. On the **Products** tab, request **“Share on LinkedIn”** and
   **“Sign In with LinkedIn using OpenID Connect”**. These grant the
   `openid`, `profile` and `w_member_social` scopes.
3. On the **Auth** tab, copy the **Client ID** and **Client Secret**, and add this exact
   **Authorized redirect URL**:
   ```
   http://localhost:8080/api/linkedin/callback
   ```
4. Put the credentials in your `.env`:
   ```bash
   LPE_LINKEDIN_CLIENT_ID=your-client-id
   LPE_LINKEDIN_CLIENT_SECRET=your-client-secret
   ```
5. Start the app with `make run`, open **http://localhost:8080**, click **Connect LinkedIn**, and
   authorize. You can then **Publish to LinkedIn** immediately or **Schedule** a post for later.

> ⚠️ Use the single-binary mode (`make run`, port **8080**) to connect LinkedIn — the OAuth
> callback is served on `8080`, so the round-trip lands there. The dev server (port 5173) is for
> UI iteration only.

> 🔐 **Token security:** OAuth tokens are encrypted at rest (AES-256-GCM) with a key auto-generated
> on first run and stored with `0600` permissions (`LPE_KEY_FILE`). Scheduled posts that fail (e.g.
> a transient API error) are retried automatically with exponential backoff before being marked
> failed.

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
