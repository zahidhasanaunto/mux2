# mux2 — master plan

## Purpose

Build an iTerm2-like terminal host on top of tmux, with an AI chat layer and an Android remote-control client, starting from a minimal verifiable loop and expanding toward the full iTerm2 AI Chat + Buddy model.

This plan is the source of truth for what gets built, in what order, and what "done" means for each task. Every task below has acceptance criteria and test cases. Milestone plan files under [docs/](docs/) break each milestone into smaller tasks with the same level of detail.

## Chosen stack

- **Backend language:** Go.
- **API hosting model:** Cloudflare Worker as the HTTP API front door; a Go host agent on the machine owns and drives tmux. The Worker is the API surface; the agent is the tmux integration point.
- **tmux integration:** host owns tmux sessions (create/kill/resize) and can attach to/drive existing user tmux sessions. Full window/pane tree model.
- **LLM:** OpenRouter for v1; any model via OpenRouter.
- **Android client:** Kotlin + Jetpack Compose.
- **Desktop verification client:** basic GUI app, developed alongside the backend.
- **v1 constraints:** no relay, no images, no local model.

## Why this stack, briefly

- Go is the best fit for the API + tmux control mode host agent in this deployment model: good tmux control-mode parsing story, easy HTTP server, easy to keep growing into the AI layer, and a clean fit for a small agent process that owns tmux on the machine.
- The Cloudflare Worker is the HTTP API front door because you chose that. A Worker cannot run tmux itself, so the architecture is Worker-first API + a Go host agent on the machine that owns tmux. We can build the Go agent first and add the Worker as the public API later, or set up wrangler when you're ready; either way the tmux ownership lives in the agent, not the Worker.
- OpenRouter gives you "any model" for v1 without committing to one vendor's SDK.
- Kotlin/Compose is the right long-term Android stack because the eventual goal is full native terminal control with real gesture support, and Compose is the modern native UI path with the deepest Android gesture/touch/pen surface.

## Architecture, v1

```
[Android app] --HTTPS/LAN--> [Go host agent on machine]
[Desktop GUI]  --HTTPS/LAN--> [Go host agent on machine]
                                              |
                                              v
                                           [tmux server]
                                              |
                               sessions / windows / panes (full tree; windows can be linked into multiple sessions)

[Chat broker + persistence] lives in the Go host agent for v1 (in-memory + SQLite).
[OpenRouter LLM client] lives in the Go host agent for v1.
[AI tool surface] maps LLM tool calls to agent operations on tmux sessions/panes.
[Chat transport to phone] is an explicit API surface in the agent for v1 (SSE or WebSocket; see M3.6).
```

For v1, all remote clients reach the Go host agent directly over LAN (HTTPS on the agent's bound port). The agent is the brain: it owns tmux, owns the chat broker and persistence, owns the OpenRouter LLM client, and exposes the HTTP + chat-transport surfaces that the desktop and Android clients call. There is **no Cloudflare Worker in v1**.

**Cloudflare Worker is deferred to M4.** In v1, the agent binds to loopback by default and only exposes LAN access after an explicit transport-security decision and an auth baseline (see M1.7 and M1.8). A Worker can become the public-facing front door later, but only after we define the network path (for example a Tunnel / private-network access) and the relay/consent model that makes "from anywhere" safe. Until then, a Worker cannot be the v1 LAN path.

## Network and security decisions, v1

- **Default bind:** loopback only (`127.0.0.1`), unless the operator explicitly enables LAN access.
- **LAN enablement:** opt-in, with an explicit transport-security decision (TLS vs. plain HTTP on a trusted LAN, and what that means for the deployment). Record the decision when made; do not assume plain HTTP over LAN is acceptable for a machine that accepts input injection, session creation, and session deletion.
- **Auth baseline:** before LAN access is enabled, require authenticated access to sensitive operations (session create/kill, input injection, resize, capture that may expose terminal contents, chat/LLM operations). The v1 auth mechanism is chosen explicitly; it can be a simple token/key sent on each request, or another mechanism, but it must exist before LAN is on.
- **Loopback default rationale:** v1 exposes input injection, session creation, and session deletion. Leaving that open on LAN by default is not acceptable without auth and a transport decision.

## Milestone overview

- **M1 — tmux control host.** Go agent drives tmux via control mode; HTTP API exposes sessions/panes/keys/paste/resize/capture; basic GUI can exercise it end to end.
- **M2 — screen + input for the phone.** Snapshot-based screen from the agent; paste + send-keys + resize over the API; basic Android screen + input client.
- **M3 — chat + OpenRouter + AI tool surface.** Chat broker, LLM client with streaming + function-call loop, tool surface for v1, execute_command Ask gating on the phone.
- **M4 — relay / "from anywhere".** Later.
- **M5 — richer iTerm2-style features.** Later, incremental.

## How to read the per-milestone plans

Each milestone has its own plan file:

- [docs/M1-plan.md](docs/M1-plan.md)
- [docs/M2-plan.md](docs/M2-plan.md)
- [docs/M3-plan.md](docs/M3-plan.md)
- [docs/M4-plan.md](docs/M4-plan.md)
- [docs/M5-plan.md](docs/M5-plan.md)

Each file breaks the milestone into tasks. Each task has:

- **Goal** — what the task produces.
- **Acceptance criteria** — the concrete "done" conditions.
- **Test cases** — how to verify, including automated tests where applicable.
- **Out of scope** — what this task deliberately does not cover.
- **Depends on** — upstream tasks or milestones.

State is tracked with checkboxes in the plan files. A task is done when its acceptance criteria are met and its test cases pass.

## Verification discipline

- Every task lists test cases. Where a task is logic-heavy (control-mode parsing, session tree, API handlers, chat broker, LLM client, tool dispatch, permission gating, chat transport), there are unit tests.
- Where a task touches tmux, there is an integration test against a real tmux server when possible, plus a documented "how to manually verify" path for things that need a real terminal interaction.
- The desktop GUI is the day-to-day verification harness for the backend before the Android app exists.
- Failure-mode tests are explicit: busy panes, delayed output, timeouts, duplicate approvals, restart-during-execution, unauthorized requests, reconnects, and replay/dedup where relevant.
- No task is marked done if its acceptance criteria are not met.

## Non-goals for v1

- Relay / "from anywhere" transport.
- Inline images.
- Local model support.
- Full iTerm2 feature parity: built-in browser, Apple Intelligence, the full 24-case remote-command catalog with man pages and file creation, shell-integration marks, orchestration watchers, HEVC/H.264 stream, signed companion consent plugin, APNs push.
- macOS-specific pieces: anything that depends on AppKit/Metal/IOKit/Sparkle.

Those are future/optional. v1 is the loop: control tmux from a host, show a screen, send input, drive it with chat + one LLM, with Ask gating for execute_command.

## Open decisions still to make

These are recorded so we don't silently make them later:

1. **Cloudflare Worker placement.** Deferred to M4. v1 is LAN: Android/desktop → Go agent. A Worker can become the public-facing front door later, after we define the network path (Tunnel / private-network access) and the relay/consent model. It is not a v1 LAN path.
2. **Snapshot screen vs. stream for M2.** v1 starts with snapshot-based screen from the agent. The stream encoder (change-driven) is a later enhancement once the loop is real.
3. **One pane per "mux2 session" vs. full tmux tree exposure.** v1 exposes the full tmux window/pane tree, including the fact that a window can be linked into multiple sessions. The UX can still present one primary pane per mux2 session for simplicity while still being able to see and drive the tree.
4. **Chat persistence backend.** SQLite in the Go agent for v1. If you want something else later, we can change it; the broker interface is what matters.
5. **OpenRouter model choice for day one.** Any model is available; we pick one for v1 testing and can change it. The plan does not pick a specific model yet.
6. **v1 auth mechanism.** Before LAN is enabled, the agent requires authenticated access. The exact mechanism (static token, per-client key, mTLS, etc.) and the transport-security decision (TLS vs. plain HTTP on a trusted LAN) are made explicitly in M1.7/M1.8, not left open.
