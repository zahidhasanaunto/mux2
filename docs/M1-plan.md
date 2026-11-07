# M1 — tmux control host

**Goal:** a Go host agent that owns and drives tmux via control mode, with an HTTP API that exposes sessions, windows, panes, send-keys, paste, resize, and pane capture; plus a basic GUI that can exercise it end to end.

**Out of scope for M1:** chat, LLM, AI tool surface, phone client, stream encoder, relay, images.

**Depends on:** nothing upstream. This is the first buildable milestone.

**Verification baseline:** the desktop GUI (or curl/python against the agent HTTP API) is the day-to-day harness. There must also be unit tests for the control-mode parsing and session-tree logic, and integration tests against a real tmux server where possible.

---

## M1.1 — project scaffold: Go host agent repo structure

**Goal:** a Go module under `backend/` with a clear package layout, a CI-ready `go test` baseline, and a README for the backend.

**Acceptance criteria:**

- [ ] `backend/` contains a Go module with a sensible layout: e.g. `cmd/mux2-agent/`, `internal/` packages, and package docs.
- [ ] `go test ./...` runs from the repo root and passes.
- [ ] There is a documented way to run the agent locally for development.
- [ ] The backend README records the host prerequisites: tmux version, where the agent expects to find or create the tmux server, and any runtime assumptions.

**Test cases:**

- [ ] `go test ./...` passes in a clean run.
- [ ] A new contributor can clone, `go build ./cmd/mux2-agent`, and run the binary with `--help` (or equivalent) without errors.

**Out of scope:** tmux control mode itself, API, GUI. This task is only the scaffold.

**Depends on:** nothing.

---

## M1.2 — tmux control-mode session ownership: create, list, kill

**Goal:** the agent can start a tmux server/session, list sessions, and kill a session, using tmux control mode.

**Acceptance criteria:**

- [ ] The agent can create a new tmux session with a given name and a start command, detached, and return a stable session identifier that maps to the tmux session.
- [ ] The agent can list sessions it owns or can see on the tmux server, returning session id, name, state, and creation time where available.
- [ ] The agent can kill a session by its stable identifier.
- [ ] Session identifiers are stable across agent restarts when they map to a real tmux session (i.e. the agent can re-resolve existing tmux sessions to its identifiers on startup).
- [ ] The agent does not lose track of sessions it created after a restart unless the tmux session itself is gone.

**Test cases:**

- [ ] Unit test for session identification and mapping logic.
- [ ] Integration test against a real tmux server: create session, list, assert the session appears, kill it, assert it is gone.
- [ ] Integration test for restart recovery: create a session, stop and restart the agent, assert the session is still discoverable and correctly mapped.
- [ ] Negative test: kill a nonexistent session returns a clear error, not a panic or silent success.

**Out of scope:**

- [ ] Windows/panes inside the session.
- [ ] attach/detach UX for humans.
- [ ] Any chat/LLM/AI layer.

**Depends on:** M1.1.

**Decision record:** the session↔window relationship is modeled as links, not strict ownership. A window can be linked into more than one session in tmux, so the tree model must allow a window to appear in multiple sessions. Internal consistency is: a pane belongs to exactly one window, a window is identified independently of sessions, and a session lists the windows linked into it.

**Probe notes (tmux 3.6 on this host):** `tmux list-sessions -F '#{session_name}\t#{session_id}\t#{session_created}\t#{session_windows}\t#{session_height}x#{session_width}'` already exposes the fields the list endpoint can return. `session_created` is epoch seconds. When a window is linked into two sessions via `link-window`, the same window identity appears in both sessions' window lists and each session has its own active pane. That shared-window case must be represented correctly in the tree model and verified in M1.3.

**Implementation note:** the initial M1.2 package is a CLI-based ownership stub behind `internal/tmux`; the real implementation is a tmux control-mode client that holds a long-lived control socket and parses the control stream. The plan tests assert behavior, not this implementation, so the internals can change.

---

## M1.3 — window/pane tree model and listing

**Goal:** the agent models and lists windows and panes inside a session, so the API can expose the full tree, not just a single pane.

**Acceptance criteria:**

- [ ] The agent can list windows in a session, with each window's id, index, name, and active-pane id.
- [ ] The agent can list panes in a window, with each pane's id, index, size, active state, and the session/window they belong to.
- [ ] The agent can expose a session tree: session → windows → panes, in a stable shape suitable for an API response.
- [ ] The agent can identify the currently active pane in a session/window.
- [ ] The tree model is internally consistent: a pane belongs to exactly one window; a window has an identity independent of sessions; a session lists the windows linked into it, and the same window identity can appear in more than one session's linked-window list (shared window).

**Test cases:**

- [ ] Unit test for the in-memory tree model: create a synthetic tree, assert parent/child relationships and active-pane selection.
- [ ] Integration test against a real tmux server: create a session, split panes, list windows, list panes, assert the tree matches the actual tmux layout.
- [ ] Integration test: rename a window, assert the listing reflects the new name.
- [ ] **Shared-window test:** link a window into two sessions (where tmux allows it), list both sessions, and assert the shared window appears in both sessions' window lists while still being the same window identity. This validates the link model directly.
- [ ] Negative test: listing a nonexistent session returns a clear error.

**Out of scope:**

- [ ] Sending input to panes.
- [ ] Capturing pane contents.
- [ ] Resizing panes.

**Probe notes (tmux 3.6 on this host):** `tmux list-windows -F '#{window_id} #{window_index} #{window_name} active-pane=#{pane_id}'` and `tmux list-panes -F '#{pane_id} #{pane_index} #{pane_width}x#{pane_height} active=#{pane_active}'` give the data the tree model needs. Window identity (`#{window_id}`) is independent of sessions; the same window can appear in more than one session after `link-window`.

---

## M1.4 — send keys and paste into a pane

**Goal:** the agent can send literal keystrokes and paste text into a pane through tmux control mode.

**Acceptance criteria:**

- [ ] The agent can send-keys into a pane by pane id, with a string of literal text.
- [ ] The agent can paste text into a pane by pane id, with a string, using the tmux paste mechanism so that newlines/submissions behave as a real paste rather than a per-character type-in where appropriate.
- [ ] The agent can distinguish "type these keystrokes" from "paste this block of text" at the API level.
- [ ] Sending input to a nonexistent or dead pane returns a clear error.
- [ ] Input is delivered to the live pane's PTY, not lost or silently dropped.

**Test cases:**

- [ ] Integration test against a real tmux server running a program that echoes input (for example a simple cat or a script that prints what it receives), assert that sent keys/pasted text appear.
- [ ] Integration test: send a command + newline and observe the shell runs it, by capturing output afterward (reuses M1.5 capture).
- [ ] Unit test for input normalization/escaping at the agent boundary (for example newline handling, any control characters that need care at the API boundary).
- [ ] Negative test: sending to a dead pane returns an error and does not corrupt another pane.

**Out of scope:**

- [ ] Capturing screen or scrollback.
- [ ] Resizing.
- [ ] Any AI or chat layer.

**Depends on:** M1.3.

---

## M1.5 — pane capture: visible screen and scrollback

**Goal:** the agent can capture the current visible screen and, where possible, scrollback for a pane, so later screen-display and AI screen-reading features have a data source.

**Acceptance criteria:**

- [ ] The agent can capture the current visible screen of a pane as text, with a defined row/column grid shape.
- [ ] The agent can capture scrollback text for a pane, with a defined maximum and offset/limit parameters.
- [ ] The capture result includes enough structure to reconstruct a grid: rows, columns, and per-cell text (and, at minimum, the grid dimensions).
- [ ] Capture of a dead or nonexistent pane returns a clear error.
- [ ] Capture is stable and repeatable enough to be used by a screen-display client and by an AI "read the screen" tool later.

**Test cases:**

- [ ] Integration test against a real tmux server: put known text on the screen, capture visible, assert the expected text/grid appears.
- [ ] Integration test: run a command that produces output, capture scrollback, assert the output appears in the captured text within the defined limits.
- [ ] Unit test for the capture text model: grid construction/parsing/round-trip.
- [ ] Negative test: capture on a dead pane returns an error.

**Out of scope:**

- [ ] Rendered pixels / images.
- [ ] A live stream of changes.
- [ ] Alternate-screen special handling beyond the basic capture semantics tmux gives.

**Depends on:** M1.3, M1.4.

---

## M1.6 — resize pane and window

**Goal:** the agent can resize a pane and/or window, so the phone and desktop clients can change terminal geometry.

**Acceptance criteria:**

- [ ] The agent can resize a pane by pane id to a given size or by adjustment.
- [ ] The agent can resize a window or its panes in a defined way through tmux control mode.
- [ ] After a resize, subsequent capture reflects the new geometry.
- [ ] Resizing a nonexistent or dead pane/window returns a clear error.
- [ ] The API exposes resize in a way the phone/desktop clients can use.

**Test cases:**

- [ ] Integration test against a real tmux server: resize a pane, capture, assert the grid dimensions changed as expected.
- [ ] Integration test: resize and then send input, assert input still works in the resized pane.
- [ ] Unit test for resize request validation (negative sizes, out-of-range values, etc.).- [ ] Negative test: resize a dead pane returns an error.

**Probe notes (tmux 3.6 on this host):** `tmux resize-pane -t <target> -x <width> -y <height>` works and `tmux list-panes -F '#{pane_width}x#{pane_height}'` reports the resulting geometry. Geometry is in tmux internal units; the capture model in M1.5 must be consistent with whatever units/resolution the pane capture uses, and the plan should record the chosen normalization explicitly.

**Out of scope:**

- [ ] Automatic layout management beyond what tmux control mode provides.
- [ ] Any AI or chat layer.

**Depends on:** M1.3, M1.5.

---

## M1.7
 — HTTP API surface for the agent, bind and access model

**Goal:** the agent exposes an HTTP API that covers the M1 operations, with an explicit bind/access model: loopback by default, LAN only after an explicit transport-security decision and an auth baseline.

**Acceptance criteria:**

- [ ] The API exposes endpoints for: list sessions, create session, kill session, list windows, list panes, send keys, paste text, capture pane (visible + scrollback), resize pane/window.
- [ ] Each endpoint has a documented request/response shape.
- [ ] The API returns clear error bodies for bad session/pane ids, dead panes, and validation errors.
- [ ] By default the agent binds to loopback only (`127.0.0.1`). LAN access is opt-in and disabled until the operator explicitly enables it.
- [ ] Before LAN access is enabled, the API requires authenticated access. The v1 auth mechanism is chosen explicitly and documented (for example a static token/key sent on each request, or another mechanism). There is no "open on LAN by default" mode.
- [ ] Before LAN access is enabled, there is an explicit transport-security decision recorded (TLS vs. plain HTTP on a trusted LAN), and the chosen mode is documented in the backend README. If plain HTTP is chosen, the conditions under which that is acceptable are stated explicitly.
- [ ] There is basic API surface documentation in the backend README, including the bind model, auth requirement, and transport decision.

**Test cases:**

- [ ] Unit/integration tests for each endpoint against a real tmux-backed agent where possible.
- [ ] Contract tests for request/response shapes: valid requests produce the documented response shape; invalid requests produce the documented error shape.
- [ ] Negative tests: nonexistent session/pane ids, bad payloads, oversized payloads where relevant.
- [ ] **Unauthorized-request tests:** requests to sensitive endpoints without valid auth are rejected with a clear error. This applies to the operations that can change the machine's terminal state or expose its contents (session create/kill, send keys, paste, resize, capture, and any chat/LLM endpoints when they exist). Read-only harmless endpoints may have a different rule, but that rule is chosen explicitly, not by accident.
- [ ] **Loopback-default tests:** by default the agent does not accept LAN connections; only loopback is reachable. When LAN is enabled, the auth and transport decisions above are enforced.
- [ ] A scripted smoke test that creates a session, sends keys, captures output, resizes, and tears down, entirely via HTTP, using valid auth.

**Out of scope:**

- [ ] Authentication/authorization hardening for v1 beyond whatever minimal protection you want on LAN. Record any decision here when made.
- [ ] WebSocket/streaming endpoints. This milestone is request/response HTTP.
- [ ] Chat/LLM/AI.

**Depends on:** M1.2–M1.6.

---

## M1.8 — basic GUI verification client, with the same access model

**Goal:** a basic GUI app that can list sessions, create a session, send keys/paste, capture and display the screen, and resize — enough to verify the backend without the Android app — using the same bind/auth/transport model as the API.

**Acceptance criteria:**

- [ ] The GUI can connect to the agent HTTP API on loopback by default, and over LAN only when LAN is enabled and auth/transport are configured.
- [ ] The GUI handles the v1 auth mechanism (for example a token/config entry) rather than assuming open access.
- [ ] The GUI can list sessions and their windows/panes.
- [ ] The GUI can create a session and show it in the list.
- [ ] The GUI can send keys and paste text into a selected pane.
- [ ] The GUI can capture and display the current screen text of a selected pane.
- [ ] The GUI can resize a selected pane/window.
- [ ] The GUI is good enough to verify M1 end to end; it does not need to be polished.

**Test cases:**

- [ ] Manual end-to-end verification using the GUI: create session, send a command, capture output, resize, kill session — all visible in the GUI.
- [ ] At least one automated or scriptable verification path for the GUI's API calls (for example shared API client code used by both the GUI and the scripted smoke test), so the GUI isn't the only verification path.
- [ ] Negative verification: GUI shows clear errors for bad session/pane ids and dead panes.

**Out of scope:**

- [ ] Chat, LLM, AI.
- [ ] Phone-style gestures.
- [ ] A polished or feature-complete terminal UI.

**Depends on:** M1.7.

---

## M1.9 — M1 acceptance test suite and release criteria

**Goal:** a focused test suite that gates M1 done-ness, plus explicit release criteria.

**Acceptance criteria:**

- [ ] There is a test suite covering the core tmux ownership, tree, send/paste, capture, resize, and API endpoints.
- [ ] The suite includes both unit tests and real-tmux integration tests where possible.
- [ ] The desktop GUI smoke test passes against a real agent + tmux.
- [ ] The M1 release criteria are met: create/list/kill sessions, list windows/panes, send keys, paste, capture visible + scrollback, resize, all reachable via HTTP and verifiable from the GUI.

**Test cases:**

- [ ] Run the full M1 test suite; it passes.
- [ ] Run the scripted HTTP smoke test; it passes.
- [ ] Manual GUI verification passes for the full create→send→capture→resize→kill flow.

**Out of scope:**

- [ ] Chat, LLM, AI, phone client, stream, relay, images.

**Depends on:** M1.1–M1.8.

---

## M1 done-ness summary

M1 is done when:

- the Go agent creates, lists, and kills tmux sessions with stable identifiers and restart recovery;
- the agent lists the full window/pane tree and the active pane;
- the agent can send keys and paste text into a pane;
- the agent can capture visible screen and scrollback for a pane;
- the agent can resize a pane/window;
- all of the above are exposed via HTTP with documented request/response/error shapes;
- the desktop GUI can exercise the full flow end to end;
- the M1 test suite passes.

When M1 is done, the backend can control a real tmux server through a documented HTTP API and a basic GUI can verify it. That is the substrate for M2 and M3.
