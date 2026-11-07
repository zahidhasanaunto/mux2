# Roadmap

Derived from [PLAN.md](PLAN.md) and the per-milestone plans.

## v1 (this plan)

- [M1](docs/M1-plan.md) — tmux control host: Go agent + HTTP API + basic GUI, with loopback-default bind, opt-in LAN, auth before LAN, and explicit transport decision.
- [M2](docs/M2-plan.md) — phone can view a session's screen and send input over LAN (direct to the agent; no Worker in v1).
- [M3](docs/M3-plan.md) — chat + OpenRouter + AI tool surface + terminal-effect permission gating + command-completion semantics + approval integrity + chat transport contract.
- M4 — relay / "from anywhere", and likely the Cloudflare Worker front door if we want a public-facing API later. Planned, not started.
- [M5](docs/M5-plan.md) — richer iTerm2-style features, incremental.

## Immediate next step

Start M1.1 (Go host agent scaffold) and M1.2 (tmux control-mode session ownership), since M1 is the substrate for everything else.

## Standing decisions recorded in the plan

See [PLAN.md](PLAN.md) → "Open decisions still to make." The main ones:

1. Cloudflare Worker placement. Deferred to M4. v1 LAN is Android/desktop → Go agent directly. A Worker can become the public-facing front door later, after we define the network path (Tunnel / private-network access) and the relay/consent model. It is not a v1 LAN path.
2. Snapshot screen vs. stream for M2? Snapshot first, stream later.
3. One pane per mux2 session vs. full tmux tree? Full tree exposed, including shared windows linked into multiple sessions. UX can still present one primary pane per session.
4. Chat persistence backend? SQLite in the agent for v1.
5. OpenRouter model choice for day one? Not picked yet; any model is available.
6. v1 auth + transport decision. Loopback default; LAN opt-in only after auth and a recorded transport-security decision.

## Hard constraints

- No relay, no images, no local model in v1.
- Desktop GUI is for verification/development, not the final product.
- Android is the real remote client.
