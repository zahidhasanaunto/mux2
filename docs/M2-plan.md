# M2 — screen + input for the phone

**Goal:** an Android client that can connect over LAN, list sessions, view a snapshot-based screen of a pane, and drive it with paste, send-keys, and resize — using the M1 agent API.

**Out of scope for M2:** chat, LLM, AI tool surface, live change-driven stream, relay, images, gesture polish beyond what's needed to prove the loop.

**Depends on:** M1 (agent + HTTP API + basic GUI).

**Verification baseline:** the Android app against the agent on LAN is the primary M2 verification; the M1 desktop GUI remains the backend verification path. There should also be automated tests for anything logic-heavy on the Android side (API client, screen model, input encoding), and manual verification for the real terminal interaction.

---

## M2.1 — Android project scaffold: Kotlin + Compose

**Goal:** an Android project under `android/` with Kotlin + Jetpack Compose, a local HTTP API client, and a runnable shell that can connect to the agent.

**Acceptance criteria:**

- [ ] `android/` contains a Kotlin/Compose Android project that builds and runs on the available SDK/NDK/Java toolchain.
- [ ] The project has a clean API client layer for the M1 agent HTTP API, in pure Kotlin, separate from UI.
- [ ] There is a documented way to run the app on a device/emulator and point it at the agent.
- [ ] `./gradlew test` (or the project's equivalent) runs and passes for the non-UI testable code.

**Test cases:**

- [ ] Build passes for the Android project.
- [ ] Unit tests for the API client layer pass (request/response shapes, error mapping, URL/config handling).
- [ ] A documented manual step exists to run the app and confirm it can reach the agent's HTTP API (even if only a connectivity check at this stage).

**Out of scope:**

- [ ] Any screen rendering or input driving yet.
- [ ] Chat, LLM, AI.

**Depends on:** M1.7 (the HTTP API that the client will call).

---

## M2.2 — Android API client for M1 operations

**Goal:** a complete Android API client covering the M1 operations the phone needs: list sessions, list windows/panes, create session, kill session, send keys, paste, capture pane, resize.

**Acceptance criteria:**

- [ ] The client can list sessions and their windows/panes.
- [ ] The client can create a session and kill a session.
- [ ] The client can send keys and paste text into a selected pane.
- [ ] The client can capture a pane's visible screen and scrollback.
- [ ] The client can resize a pane/window.
- [ ] The client maps API errors into clear app-level errors, not raw failures that crash the UI.
- [ ] The client is testable in isolation from the UI (for example via an interface/abstraction over HTTP).

**Test cases:**

- [ ] Unit tests for the client against recorded/mocked API responses for each operation, asserting correct request shapes and response parsing.
- [ ] Error-path tests: nonexistent session/pane, dead pane, bad payload, network error — all produce app-level errors the UI can show.
- [ ] Contract tests that the client's request shapes match the M1 API contract.

**Out of scope:**

- [ ] Any UI.
- [ ] Chat, LLM, AI.

**Depends on:** M2.1, M1.7.

---

## M2.3 — snapshot-based screen viewing on Android

**Goal:** the Android app can show the current screen text of a selected pane in a readable terminal-like display, refreshed on demand.

**Acceptance criteria:**

- [ ] The app can select a session/window/pane and display its current visible screen as text in a basic terminal-like UI.
- [ ] The display reflects the grid shape well enough to be readable (rows/columns, monospace rendering, basic handling of the text grid from M1 capture).
- [ ] The user can refresh the screen manually.
- [ ] The screen view updates when the user navigates to a different pane.
- [ ] The screen view handles empty/dead panes gracefully (clear error or empty state, not a crash).

**Test cases:**

- [ ] Manual verification: create a session, run a command with known output, refresh the screen view, assert the output is visible.
- [ ] Manual verification: split panes, switch which pane is shown, assert the view changes.
- [ ] Unit/UI tests for the screen rendering piece where practical (for example rendering a synthetic text grid into the composable and asserting visible structure).
- [ ] Negative verification: dead/empty pane shows a clear state.

**Out of scope:**

- [ ] Live change-driven stream.
- [ ] Images/in-line media.
- [ ] Scrollback browsing UI beyond what's needed to prove the loop (scrollback can be available via capture but the M2 UI may be "current screen" only).

**Depends on:** M2.2, M1.5.

---

## M2.4 — input driving from the phone: send keys and paste

**Goal:** the Android app can type text and paste text into the selected pane, and the result is visible in the screen view.

**Acceptance criteria:**

- [ ] The app can send keys into the selected pane and the input reaches the live terminal.
- [ ] The app can paste text into the selected pane and the pasted block reaches the live terminal.
- [ ] The user can see the result of input via the screen view after sending/pasting.
- [ ] Sending/pasting to a dead or nonexistent pane shows a clear error.
- [ ] Basic input UX is usable enough to prove the loop; it does not need to be a full terminal keyboard yet.

**Test cases:**

- [ ] Manual verification: send a command + newline, refresh screen, assert the command ran and output is visible.
- [ ] Manual verification: paste a multi-line snippet, refresh screen, assert it arrived.
- [ ] Negative verification: sending/pasting to a dead pane shows an error.
- [ ] Unit tests for any input encoding/normalization at the app boundary (newline/CR handling, any control characters that need care).

**Out of scope:**

- [ ] A full soft keyboard with terminal control key handling (that can come later; M2 only needs enough to prove send-keys and paste work).
- [ ] Chat, LLM, AI.

**Depends on:** M2.2, M1.4.

---

## M2.5 — resize from the phone

**Goal:** the Android app can resize the selected pane/window and see the changed geometry in the screen view.

**Acceptance criteria:**

- [ ] The app can resize a selected pane/window.
- [ ] After resize, the screen view reflects the new geometry.
- [ ] Resizing a dead or nonexistent pane/window shows a clear error.
- [ ] The resize UX is usable enough to prove the flow.

**Test cases:**

- [ ] Manual verification: resize a pane, refresh screen, assert the grid dimensions changed.
- [ ] Manual verification: resize, then send input, assert input still works.
- [ ] Negative verification: resize a dead pane shows an error.

**Out of scope:**

- [ ] Automatic layout management.
- [ ] Chat, LLM, AI.

**Depends on:** M2.2, M1.6.

---

## M2.6 — M2 acceptance test suite and release criteria

**Goal:** a focused verification suite for M2 and explicit release criteria.

**Acceptance criteria:**

- [ ] The Android API client has unit tests covering all operations and error paths.
- [ ] The screen view and input pieces have tests where practical.
- [ ] Manual end-to-end verification on a real device/emulator against the agent on LAN passes: list sessions, create one, send a command, see output, paste text, see it, resize, see new geometry.
- [ ] The M2 release criteria are met: the phone can connect over LAN, view a session's screen, and drive it with send-keys, paste, and resize.

**Test cases:**

- [ ] Run the Android client unit tests; they pass.
- [ ] Run the build; it passes.
- [ ] Manual end-to-end on device/emulator against the agent; the full flow works.

**Out of scope:**

- [ ] Chat, LLM, AI, stream, relay, images.

**Depends on:** M2.1–M2.5.

---

## M2 done-ness summary

M2 is done when:

- the Android app builds and runs on the available toolchain;
- the Android API client covers list/create/kill sessions, list windows/panes, send keys, paste, capture, resize, with unit tests and error mapping;
- the app can show a selected pane's current screen as text in a readable terminal-like UI;
- the app can send keys and paste text into the selected pane and see the result;
- the app can resize a pane/window and see the new geometry;
- the full flow is verified manually on a real device/emulator against the agent on LAN.

When M2 is done, you can control a real tmux-backed terminal from the phone over LAN: view the screen and send input. That is the substrate for M3's chat + AI layer.
