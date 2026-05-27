# TODO / Future work

Items intentionally deferred. The app is fully functional for **personal/local
use** today (build with `make build`, run with `make run`); nothing here is
required to use it.

## Release & distribution (deferred)

Goal: let anyone download and run the app **without** a Go/Node toolchain —
delivering on the project's "single binary, zero friction" promise (ADR-001).

- [ ] **Cross-platform binaries** via a tag-triggered GitHub Actions workflow
      (`v*`), most likely with **GoReleaser**:
      - targets: Linux, macOS, Windows × amd64/arm64 (SQLite is CGO-free, so
        cross-compilation is cheap — ADR-002);
      - attach binaries + `SHA256` checksums to a GitHub Release;
      - generate the changelog from the Conventional Commits already in use.
- [ ] **Docker image** — multi-stage `Dockerfile` (Node build → Go build →
      minimal `scratch`/`distroless` final), published to GHCR, enabling
      `docker run` with `data.db` on a volume (see the deployment diagram).
- [ ] **Version stamping** — embed version/commit via `-ldflags` and expose
      them in `GET /api/health`.
- [ ] **Install docs** — README section: "download the binary for your OS" and
      `docker run`.

Notes:
- Stays zero-cost: GitHub Actions + Releases + GHCR are free for a public repo.
- This is *distribution*, not cloud hosting — the product remains local-first.
- Can be validated locally before any real release via
  `goreleaser release --snapshot` and a local `docker build`.

## Other ideas

- [ ] Local metrics over time (e.g. posts per week), beyond the current totals.
- [ ] Manual re-queue button for posts marked `failed` in the UI.
