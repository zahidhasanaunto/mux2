# mux2 desktop verification client

A basic GUI app that exercises the mux2 Go host agent over its HTTP API, so the
backend can be verified end to end before the Android app exists.

## Purpose

This client is for verification and development, not the final product. v1 remote
control lives in the Android app (`../android/`). The desktop client exists so M1
can be checked manually and so shared API plumbing can be exercised from a real UI.

## Scope

- Connect to the agent HTTP API with the v1 auth model (bearer API key), using
  loopback by default and LAN only when LAN is enabled.
- List sessions and their windows/panes.
- Create a session (with a start command).
- Kill a session.
- Send keys and paste text into a selected pane.
- Capture and display the current screen text of a selected pane.
- Resize a selected pane/window.

## Stack

The exact stack is intentionally left as a decision to make when M1.8 is started.
This README records the constraint and the candidates; it does not commit to one
yet.

Constraints:
- Should be buildable/runable in this environment for v1 verification.
- Should share as much API plumbing as practical with any scripted smoke test path,
  so the GUI is not the only verification path.
- Prefer something that can be iterated quickly alongside the Go backend.

Candidates, not decided yet:
- Go + a lightweight GUI toolkit, sharing types with the backend via a shared
  module or duplicated API types.
- Node + a web-based desktop shell (for example nw.js / Electron / Tauri-style),
  with the API client in JS/TS.
- Python + PyGObject/GTK, since pygobject is present in this environment.

The chosen stack is recorded in this README when M1.8 starts.

## v1 constraints

- No relay, no images, no local model.
- Uses the same bind/auth/transport model as the agent: loopback by default,
  LAN opt-in, auth before LAN.
- Plain HTTP on a trusted LAN for v1, with auth required before LAN is enabled
  (see `../docs/DECISIONS.md`).

## Verification

- Manual end-to-end: create session, send a command, capture output, resize,
  kill session — all visible in the GUI.
- At least one automated/scriptable verification path for the same API calls,
  so the GUI is not the only verification path.
