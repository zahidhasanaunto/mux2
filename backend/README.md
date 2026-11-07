# mux2 backend

The Go host agent that owns and drives tmux on the machine, exposes the M1 HTTP API,
and (later) hosts the chat broker, OpenRouter LLM client, and AI tool surface.

## Stack

- Go 1.26 (module at `backend/`).
- tmux control mode on the host for session/window/pane operations.
- Loopback-only by default; LAN is opt-in (see "Bind and access model").

## Prerequisites

- tmux 3.6 on the host.
- The agent expects to find or start a tmux server on the host. By default it uses
  the user's normal tmux server socket; a separate socket can be chosen for
  isolated plan testing. See "Development server" below.
- A shell is available for starting sessions (the agent starts sessions detached
  with a start command).

## Build and run

From the repo root:

```bash
cd backend

# build
go build ./cmd/mux2-agent

# run tests
go test ./...
```

The agent binary is `cmd/mux2-agent`. The current binary exits with a notice that
the implementation is not yet in place; after M1.2 it starts driving tmux.

## Bind and access model

- **Default:** bind to loopback only (`127.0.0.1`).
- **LAN:** opt-in, off by default. The operator explicitly enables LAN access.
- **Auth:** before LAN is enabled, authenticated access is required on any endpoint
  that can change the machine's terminal state or expose its contents. The chosen
  v1 auth mechanism is a static bearer API key sent as `Authorization: Bearer <key>`
  on every request to sensitive endpoints. LAN is not enabled unless an auth key is
  configured.
- **Transport decision (v1):** plain HTTP on a trusted LAN, with auth enabled before
  LAN. This is an explicit, recorded decision; see `../docs/DECISIONS.md`.

## API surface (planned, implemented across M1.2–M1.7)

M1 HTTP endpoints (request/response shapes documented as they are implemented):

- list sessions
- create session (name + start command, detached)
- kill session
- list windows in a session
- list panes in a window
- send keys into a pane
- paste text into a pane
- capture pane (visible screen + scrollback)
- resize pane/window

All sensitive endpoints require the bearer API key. Error bodies are documented for
bad session/pane ids, dead panes, and validation errors.

## tmux server notes

- The agent drives tmux through control mode where possible.
- The session↔window relationship is modeled as links, not strict ownership: a
  window can be linked into more than one session. A pane belongs to exactly one
  window; a window has an identity independent of sessions.
- Session identifiers are stable across agent restarts when they map to a real tmux
  session; on startup the agent re-resolves existing tmux sessions to its identifiers.

## Safety notes

- v1 exposes input injection, session creation, and session deletion. Those are
  gated by auth before LAN, and LAN is off by default. Do not expose the agent on
  an untrusted network without revisiting the transport/auth model.
- No relay, no images, no local model in v1. OpenRouter is for M3 only.

## Development server (plan testing)

For plan integration tests that must not touch your main tmux server, use a
dedicated tmux socket:

```bash
tmux -L mux2-plan-test -f /dev/null start-server
# run the agent configured to use that socket if/when the agent supports it
tmux -L mux2-plan-test kill-server
```

This socket path lives under `$TMPDIR/tmux-$(id -u)` (or `/tmp`); treat it as
ephemeral for plan tests.
