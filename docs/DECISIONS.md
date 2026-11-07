# M1 decision records (v1 auth + transport)

These are the explicit choices made for M1.7/M1.8. They are recorded here so the
scaffold and API are built against a concrete model, not "choose explicitly" left
open during implementation.

## v1 auth mechanism — chosen

**Chosen:** static bearer API key, sent as `Authorization: Bearer <key>` on every
request to sensitive endpoints.

- One operator-chosen key per host.
- The key is required on all endpoints that can change the machine's terminal
  state or expose its contents: session create/kill, send keys, paste, resize,
  capture, and (later) chat/LLM endpoints.
- Read-only harmless endpoints may have a different rule only if that rule is
  chosen explicitly; the default is "auth required" on anything worth exposing.
- The CLI flag `--api-key` (or env / config) sets the key; the agent refuses to
  expose LAN without one configured for the chosen auth mechanism.

So the scaffold should already require auth on the sensitive endpoints and return
a clear `401`-style error for missing/wrong credentials.

## v1 transport-security decision — chosen

**Chosen:** plain HTTP on a trusted LAN, with the explicit conditions recorded.

- v1 does **not** terminate TLS on the agent in this first milestone.
- Plain HTTP is acceptable for v1 only under these explicit conditions:
  - The agent is on a trusted LAN (reachability is limited to the intended
    machine/phone on the same network).
  - LAN access is opt-in and off by default (loopback-only until the operator
    enables it).
  - Auth is enabled before LAN is enabled, so even on the trusted LAN the agent
    does not accept unauthenticated sensitive requests.
  - The operator accepts the trade-off that traffic on the LAN is not encrypted
    by the agent itself in v1.
- The backend README states this decision and the conditions. If later milestones
  (M4 or a TLS add-on) want TLS, that is a separate explicit decision then.

So "loopback by default, LAN opt-in, auth before LAN" is already concrete in the
plan; this record makes the plain-HTTP-on-trusted-LAN part concrete too.

## What this means for M1.1–M1.8

- M1.1 scaffold can already assume: bearer API key on sensitive endpoints,
  loopback-only bind by default, a flag/env for the API key, and a README section
  that records the chosen auth + transport decision.
- M1.7/M1.8 do **not** need to leave auth/transport open as a question; they
  implement the chosen model and its tests.

## Scope of this record

- This record covers v1 auth and transport only.
- It does not decide M4 relay/consent model, TLS-in-later-versions, per-client keys,
  mTLS, or any stronger auth model. Those remain open for later, if wanted.
