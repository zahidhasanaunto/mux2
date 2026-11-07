# mux2

A Linux terminal host with tmux as the backend, an HTTP API, an AI chat layer, and an Android remote-control client. The short version of the stack:

- **Backend language:** Go (for the tmux host agent + API surfaces).
- **API hosting model:** Cloudflare Worker as the HTTP API front door, with a Go host agent running on the machine that actually owns and drives tmux. The Worker is the API surface; the agent is the tmux integration point.
- **tmux integration:** host owns tmux sessions (create/kill/resize them) AND can attach to/drive existing user tmux sessions. Full window/pane tree model, not one-pane-per-session.
- **LLM:** OpenRouter for v1, any model via OpenRouter.
- **Android client:** Kotlin + Jetpack Compose, built for native gesture support and future full-terminal control.
- **Desktop verification client:** basic GUI app, developed alongside the backend.
- **v1 constraints:** no relay, no images, no local model.

See [docs/PLAN.md](docs/PLAN.md) for the master plan and per-milestone plan files under [docs/](docs/).

## Repository layout

- `docs/` — plan, milestones, task breakdowns, acceptance criteria, test cases.
- `backend/` — Go host agent + API surfaces.
- `desktop/` — basic GUI verification client.
- `android/` — Kotlin/Compose Android remote-control client.
- `infra/` — Cloudflare Worker / deployment scaffolding.

## v1 scope

M1–M3 is the real product for v1. M4 (relay / "from anywhere") is later. M5 is "make it feel like iTerm2's AI chat" and is optional/incremental.

- **M1** — tmux control host: Go agent drives tmux via control mode; HTTP API exposes sessions/panes/keys/paste/resize/capture; basic GUI can exercise it.
- **M2** — screen + input for the phone: snapshot-based screen from the agent, paste + send-keys + resize over the API, basic Android screen + input client.
- **M3** — chat + OpenRouter + AI tool surface: chat broker, one LLM client with streaming + function-call loop, tool surface for v1, execute_command Ask gating on the phone.
- **M4** — relay / "from anywhere" (later).
- **M5** — richer iTerm2-style features (later, incremental).

## Hard constraints for v1

- No relay in v1. LAN control only.
- No inline images in v1.
- No local model in v1. OpenRouter only.
- Desktop GUI is for verification and development, not the final product. The real remote client is the Android app.
- tmux 3.6 on the host machine for now.

## Tooling notes for this environment

- Go 1.26.8, Rust 1.99.1, Node 22.23.4, npm 10.9.9, Python 3.14.4.
- tmux 3.6 on the host.
- Cloudflare Wrangler is **not** installed in this environment yet. If we want Workers in this environment, that needs to be installed first. For now, the plan treats the Worker as the API front door and the Go host agent as the tmux owner; we can build the Go agent first and add the Worker later, or set up wrangler when you're ready.
- Android SDK present at `$ANDROID_HOME` with build-tools 35/36/37 and NDK 27.1.12297006, Java 17.
- Git remote is SSH now (`git@github.com:zahidhasanaunto/mux2.git`), upstream is `origin/master`.
