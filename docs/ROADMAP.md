# Roadmap

Derived from [PLAN.md](PLAN.md) and the per-milestone plans.

## v1 (this plan)

- [M1](docs/M1-plan.md) — tmux control host: Go agent + HTTP API + basic GUI.
- [M2](docs/M2-plan.md) — phone can view a session's screen and send input over LAN.
- [M3](docs/M3-plan.md) — chat + OpenRouter + AI tool surface + execute_command Ask gating.
- M4 — relay / "from anywhere". Planned, not started.
- [M5](docs/M5-plan.md) — richer iTerm2-style features, incremental.

## Immediate next step

Start M1.1 (Go host agent scaffold) and M1.2 (tmux control-mode session ownership), since M1 is the substrate for everything else.

## Standing decisions recorded in the plan

See [PLAN.md](PLAN.md) → "Open decisions still to make." The main ones:

1. Cloudflare Worker first or Go agent first? Plan assumes Go agent first for tmux ownership; Worker added as the public API layer later.
2. Snapshot screen vs. stream for M2? Snapshot first, stream later.
3. One pane per mux2 session vs. full tmux tree? Full tree exposed; UX can still present one primary pane per session.
4. Chat persistence backend? SQLite in the agent for v1.
5. OpenRouter model choice for day one? Not picked yet; any model is available.

## Hard constraints

- No relay, no images, no local model in v1.
- Desktop GUI is for verification/development, not the final product.
- Android is the real remote client.
