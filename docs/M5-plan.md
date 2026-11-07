# M5 — richer iTerm2-style features

**Status:** incremental, selective. M5 is where we add iTerm2-style features beyond the v1 loop, only the ones worth porting.

**Goal:** selectively expand the v1 loop toward the iTerm2 AI Chat + Buddy feature set, porting the design (not the macOS-specific code) for the features that matter.

**Out of scope for v1:** everything in M5. v1 is M1–M3 plus M4 if scheduled.

**Depends on:** M1–M3 (and M4 if we want companion-style remote features).

---

## M5.1 — expand the tool catalog selectively

**Goal:** add more LLM 도구 from the iTerm2 remote-command catalog where they make sense on Linux + tmux.

**Candidate tools (pick selectively):**

- `set_clipboard`
- `restart_session` (where tmux semantics allow a restartable session)
- `create_file`
- `get_man_page`
- `get_command_history` / `get_last_command` / `get_command_output` / `search_command_history` (better output-boundary handling, possibly a lighter shell-integration analog if we build one)
- `get_shell_type`, `get_remote_hostname`, `get_user_identity`, `detect_ssh_session`, `get_current_directory`
- `is_at_prompt`

**Acceptance criteria:**

- [ ] Each added tool has a name, description, parameter schema, dispatch mapping to agent/tmux operations, bounded result formatting, and unit + integration tests.
- [ ] Tools that need output-boundary semantics have a documented design for how boundaries are determined on Linux + tmux (since tmux lacks iTerm2's shell-integration marks).

**Test cases:**

- [ ] Unit tests per tool for argument parsing/validation.
- [ ] Integration tests against real tmux where the tool touches the terminal.
- [ ] Negative tests for missing/invalid args and dead/nonexistent sessions.

**Depends on:** M3.3.

---

## M5.2 — better screen model: alternate screen + mouse reporting awareness

**Goal:** make the screen view and the AI's screen reading aware of alternate screen and mouse reporting semantics, so TUI programs and terminal mouse mode behave correctly on the phone.

**Acceptance criteria:**

- [ ] The screen model distinguishes normal screen vs. alternate screen where tmux/agent semantics allow it.
- [ ] The phone's scroll/touch gestures can be mapped to terminal mouse reporting when the pane is in mouse mode, rather than naive back-scroll.
- [ ] The AI's `get_screen_contents` reflects the screen kind and mouse reporting state.

**Test cases:**

- [ ] Manual verification with a TUI program that uses the alternate screen (for example a pager or editor) and a mouse-aware program where possible.
- [ ] Unit tests for the screen model's kind/mouse-reporting fields.

**Depends on:** M2, M3.3.

---

## M5.3 — richer stream: change-driven encoded stream (later)

**Goal:** replace/augment snapshot-based screen with a change-driven encoded stream for the live view, so the phone's live view is efficient over worse connections.

**Acceptance criteria:**

- [ ] To be designed when scheduled: pick an encoding approach suitable for Linux (not VideoToolbox HEVC/H.264 the way iTerm2 does on macOS).
- [ ] The stream is change-driven with pacing/dedup, and snapshots remain the fallback/verification path.

**Test cases:**

- [ ] To be written when scheduled.

**Depends on:** M2, M1.5.

---

## M5.4 — companion-style remote features (later)

**Goal:** if we want the Android app to be a true iTerm2 Buddy-style remote control beyond chat + basic input, add features like selection gestures, browser input (if we ever add a built-in browser), modal alert handling on the host, and APNs-like push for watch fires / chat updates (if we add async watchers).

**Acceptance criteria:**

- [ ] To be designed when scheduled.

**Test cases:**

- [ ] To be written when scheduled.

**Depends on:** M3, M4 if remote features are desired.

---

## M5.5 — orchestration-style multi-session driving (later, optional)

**Goal:** if we want an "orchestrator" mode that can see and drive many sessions with watchers and async notifications, design and implement a selective version of that.

**Acceptance criteria:**

- [ ] To be designed when scheduled.

**Test cases:**

- [ ] To be written when scheduled.

**Depends on:** M3, M1 session model.

---

## M5 done-ness

M5 has no single done-ness criterion; each sub-task is done on its own acceptance criteria. M5 is the "make it feel more like iTerm2's AI Chat + Buddy" layer, added incrementally after the v1 loop is real.
