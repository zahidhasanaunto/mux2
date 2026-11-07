# mux2

A Linux terminal host with tmux as the backend, an HTTP API, an AI chat layer, and an Android remote-control client. The short version of the stack:

- **Backend language:** Go (for the tmux host agent + API surfaces).
- **Backend language:** Go (for the tmux host agent + API surfaces).
- **API hosting model, v1:** Android/desktop clients reach the Go host agent directly over LAN. The Cloudflare Worker is deferred to M4 (likely as the public-facing front door and/or relay-facing layer once the network path and relay/consent model are defined). A Worker cannot own tmux, so it is not the v1 tmux integration point.
- **tmux integration:** host owns tmux sessions (create/kill/resize them) AND can attach to/drive existing user tmux sessions. Full window/pane tree model, not one-pane-per-session. A window can be linked into multiple sessions; the tree model distinguishes window identity from session links/indexes.
- **LLM:** OpenRouter for v1, any model via OpenRouter.
- **Android client:** Kotlin + Jetpack Compose, built for native gesture support and future full-terminal control.
- **Desktop verification client:** basic GUI app, developed alongside the backend.
- **v1 constraints:** no relay, no images, no local model. v1 has no Cloudflare Worker in the hot path; the Worker is an M4 item.

See [docs/PLAN.md](docs/PLAN.md) for the master plan and per-milestone plan files under [docs/](docs/).

## Repository layout

- `docs/` — plan, milestones, task breakdowns, acceptance criteria, test cases.
- `backend/` — Go host agent + API surfaces.
- `desktop/` — basic GUI verification client.
- `android/` — Kotlin/Compose Android remote-control client.
- `infra/` — Cloudflare Worker / deployment scaffolding.

## v1 scope

M1–M3 is the real product for v1. M4 (relay / "from anywhere") is later. M5 is "make it feel like iTerm2's AI chat" and is optional/incremental.

- **M1** — tmux control host: Go agent drives tmux via control mode; HTTP API exposes sessions/panes/keys/paste/resize/capture; basic GUI can exercise it; loopback-default bind with opt-in LAN, auth before LAN, and an explicit transport decision.
- **M2** — screen + input for the phone: snapshot-based screen from the agent, paste + send-keys + resize over the API, basic Android screen + input client with credential entry/storage and the chosen transport config.
- **M3** — chat + OpenRouter + AI tool surface: chat broker, one LLM client with streaming + function-call loop, v1 AI tool surface with a single terminal-input permission category, explicit command-completion semantics, approval integrity, run cancellation + iteration/time limits, durable vs. best-effort event model, and a chat transport contract (subscriptions + writes) with auth.
- **M4** — relay / "from anywhere", and likely the Cloudflare Worker front door if we want a public-facing API later. Planned, not started.
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
- Cloudflare Wrangler is **not** installed in this environment yet, and M4 (where the Worker lives) is not v1. If we want Workers in this environment later, wrangler needs to be installed then. For v1, the plan is Go host agent + LAN; no Worker in the hot path.
- Android SDK present at `$ANDROID_HOME` with build-tools 35/36/37 and NDK 27.1.12297006, Java 17.
- Git remote is SSH now (`git@github.com:zahidhasanaunto/mux2.git`), upstream is `origin/master`.
