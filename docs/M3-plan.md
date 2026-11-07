# M3 — chat + OpenRouter + AI tool surface

**Goal:** add a chat broker with persistence, an OpenRouter LLM client with streaming and a function-call loop, and an AI tool surface that maps LLM tool calls to the M1/M2 tmux operations, with execute_command gated by an Ask/Allow/Deny flow on the phone.

**Out of scope for M3:** relay, images, local model, the full 24-case iTerm2 remote-command catalog, shell-integration marks, orchestration watchers, built-in browser, inline media.

**Depends on:** M1 (agent + tmux control) and M2 (phone can view screen + send input). M3 adds the chat + AI layer on top of that.

**Verification baseline:** the Android app + agent together, with the chat UI on the phone and the AI tool loop running against the real tmux session. There should be unit tests for the chat broker, persistence, LLM client, tool dispatch, and permission gating, plus manual end-to-end verification for the real loop.

---

## M3.1 — chat broker and persistence

**Goal:** a chat broker that supports multiple chats, message persistence, and pub/sub delivery, living in the Go host agent for v1.

**Acceptance criteria:**

- [ ] The broker supports creating chats and listing chats.
- [ ] The broker supports adding messages to a chat and listing messages for a chat, with persistence (SQLite in the agent for v1).
- [ ] The broker supports subscriptions: a subscriber gets existing messages on subscribe and live deliveries for new messages.
- [ ] Messages survive agent restarts (persisted), and the broker can replay history on subscribe.
- [ ] The broker has a clean interface so the LLM client, the phone, and the desktop GUI can all publish and subscribe.

**Test cases:**

- [ ] Unit tests for the broker: create chat, add messages, subscribe, assert delivery of existing + new messages.
- [ ] Unit tests for persistence: restart the broker state (or re-open the store) and assert messages survive and can be replayed.
- [ ] Concurrency tests where practical: concurrent publishes and subscriptions behave correctly.
- [ ] Negative tests: operations on nonexistent chats return clear errors.

**Out of scope:**

- [ ] Chat UI on the phone. That comes in M3.5.
- [ ] Relay or cross-machine sync. v1 broker is in the agent.

**Depends on:** M1.7 (the agent is where the broker lives; it can be reachable from the phone via the API).

---

## M3.2 — OpenRouter LLM client with streaming and function-call loop

**Goal:** an LLM client in the agent that calls OpenRouter, supports streaming deltas, and runs the function-call/response loop.

**Acceptance criteria:**

- [ ] The client can send a chat-style prompt to OpenRouter and receive a streaming response, emitting deltas as they arrive.
- [ ] The client can handle a response that includes tool calls (function calls), extract the tool call arguments, and feed tool results back for a follow-up request.
- [ ] The client can complete the loop: prompt → tool call → tool result → final response, emitting the final answer as a chat message.
- [ ] The client is configured via OpenRouter (API key, model choice, endpoint), with the model choice swappable.
- [ ] The client maps OpenRouter response shapes into the agent's internal message model cleanly.
- [ ] Streaming and non-streaming paths are both usable; v1 can default to streaming.

**Test cases:**

- [ ] Unit tests for the message model mapping between the agent and the OpenRouter wire shape.
- [ ] Unit tests for the function-call loop logic using a mock HTTP layer: prompt with tool-call response → feed tool result → final response.
- [ ] Unit tests for streaming delta accumulation into a complete message.
- [ ] Integration/contract tests against OpenRouter with a real API key where possible, for a small prompt and for a tool-call round-trip if OpenRouter supports the chosen model's tool format.
- [ ] Negative tests: API errors, malformed responses, tool-call responses without a final answer, rate-limit/type errors — all produce clear errors the rest of the system can handle.

**Out of scope:**

- [ ] Multiple LLM vendors. v1 is OpenRouter only.
- [ ] Model catalog UI or per-chat model selection UI. That can come later; for v1 the model choice can be a config value.

**Depends on:** M3.1 (the client emits messages into the broker), M1.7 (the client lives in the agent).

---

## M3.3 — AI tool surface for v1, with terminal-effect permissions

**Goal:** a set of LLM 도구 (tool/function) definitions that map LLM tool calls to the agent's tmux operations, with permissions defined around terminal effects, not only around `execute_command`.

**Acceptance criteria:**

- [ ] The tool surface includes at minimum: `execute_command`, `insert_text_at_cursor`, `get_screen_contents`, plus a few session-state/read tools such as `get_terminal_size`, `get_current_directory`, and `get_last_command` where the underlying tmux/agent semantics allow it.
- [ ] Each tool has a clear name, description, and parameter schema, and is registered with the LLM client.
- [ ] The permission model is defined around **terminal effects**, not only around "run a command". In v1, the tools that can affect the terminal are gated unless explicitly allowed:
  - `execute_command` — runs a command in the selected pane. Canonical "run command" tool.
  - `insert_text_at_cursor` — types/pastes text into the selected pane as keystrokes. This is a terminal write and can include submission/control characters, so it is gated by default in v1 unless the chat explicitly allows AI terminal writes.
  - paste/text input paths that can submit or carry control characters are treated as terminal writes under the same default-gated rule.
  - `resize` and session create/kill are also terminal-state-changing operations and are gated where they are reachable through the AI tool surface.
- [ ] `execute_command` maps to sending a command into the selected pane and capturing its output via the M1 send+capture flow, with the completion semantics defined in M3.3a below.
- [ ] `insert_text_at_cursor` maps to sending text into the selected pane as keystrokes through the M1 send-keys flow.
- [ ] `get_screen_contents` maps to capturing the pane's visible screen (and scrollback where relevant) through the M1 capture flow.
- [ ] The tool surface can be targeted at a specific session/pane (the AI's "current session"), resolved from the chat's linked session.
- [ ] Tool results are formatted so the LLM can use them (clear output, bounded size where relevant).

**Test cases:**

- [ ] Unit tests for each tool's argument parsing and validation.
- [ ] Unit tests for the tool dispatch logic: given a tool call, the dispatcher invokes the correct agent operation and returns a result.
- [ ] Integration tests against a real tmux-backed agent for the key tools: execute_command produces output, insert_text_at_cursor types text, get_screen_contents returns the screen.
- [ ] Negative tests: tool calls with missing/invalid arguments return clear errors; tools targeted at a dead/nonexistent session return clear errors.

### M3.3a — command completion semantics for `execute_command`

**Goal:** define what `execute_command` actually guarantees, because "send + capture" is not "command completed".

**Acceptance criteria:**

- [ ] `execute_command` does **not** claim completed execution unless command-boundary tracking is available for the supported shell.
- [ ] If command-boundary tracking is not available, `execute_command` returns one of two explicit result shapes:
  - **"input submitted"** plus a snapshot of the pane after submission, clearly labeled as not a completion guarantee; or
  - a bounded wait for a shell prompt/mark where supported, with the wait bounded by an explicit timeout.
- [ ] The tool documents which shells/terminals support boundary tracking and which do not, so the LLM (and the user) know what the result means.
- [ ] Slow/busy pane behavior is explicit: a slow command does not block the tool forever; a busy pane does not silently attribute input to the wrong program.

**Test cases:**

- [ ] **Slow-command test:** submit a command that takes longer than the bounded wait, assert the result is the explicit "input submitted" / timeout shape, not a fake completion.
- [ ] **Delayed-output test:** submit a command whose output appears after a delay, assert the tool does not falsely claim completion at submission time.
- [ ] **Busy-pane test:** submit input to a pane that is mid-session in another program, assert the result does not claim the input was a completed shell command.
- [ ] **Timeout test:** the bounded wait has a configurable timeout and does not hang the agent.
- [ ] **Completion-guarantee test (where supported):** where a shell/mark-based boundary is available, assert the tool can report completion and output; where it is not available, assert it does not claim completion.

**Out of scope:**

- [ ] The full 24-case iTerm2 remote-command catalog (man pages, file creation, clipboard, restart session, browser actions, SSH detection, etc.). Those are M5 if we want them.
- [ ] A full shell-integration marks layer. tmux does not provide those natively; better output-boundary handling is a design decision for M3.3a/M5, not assumed here.

**Depends on:** M3.1, M3.2, M1.4, M1.5, M2.2.

---

## M3.4 — permission gating for terminal effects

**Goal:** terminal-affecting AI tools are gated by default, with the approval flow surfaced on the phone. `execute_command` is the canonical run-command tool, but `insert_text_at_cursor`/paste (and any AI-reachable resize or session create/kill) are also terminal writes and are gated unless explicitly allowed.

**Acceptance criteria:**

- [ ] Each chat has a permission setting per terminal-effect tool category (at minimum: run command, AI terminal writes, and any other AI-reachable state-changing operations such as resize or session create/kill if they are exposed to the AI). The v1 setting set is chosen explicitly; it is not limited to `execute_command` alone.
- [ ] Each setting is one of Allow, Ask, Deny for that category.
- [ ] When the LLM requests a gated tool and the chat setting is Ask, the request is parked and the phone is asked to approve or deny before the tool runs.
- [ ] When allowed, the tool runs and the result is fed back into the chat.
- [ ] When denied, the tool result reflects the denial and the chat continues.
- [ ] The approval is addressable by an immutable pending-request identity so the phone can answer the specific pending request.
- [ ] The permission state is persisted (so a setting chosen on the phone survives restarts).

**Test cases:**

- [ ] Unit tests for the permission state machine per category: Allow runs immediately, Ask parks and waits, Deny blocks, and the transitions are correct.
- [ ] Unit tests for the pending-request resolution: answering a specific pending request runs or blocks the right tool call.
- [ ] **Bypass tests:** verify that AI terminal writes cannot bypass the run-command approval by smuggling submission/control characters through `insert_text_at_cursor` or paste when terminal writes are Ask/Deny. Attempted bypass payloads (for example text containing a command followed by a submission character) are either blocked or required to go through the gated run-command path, depending on the chosen policy, and the test asserts the chosen policy holds.
- [ ] Integration tests against the real agent + tmux for the Allow path: an approved run-command tool runs and returns output under the M3.3a completion semantics.
- [ ] Negative tests: answering a nonexistent or already-resolved request is handled cleanly.

**Out of scope:**

- [ ] A safety classifier or "Always Allow still blocked by safety" layer. That is a later enhancement if you want it; v1 Ask gating is the core safety behavior.
- [ ] Per-tool permissions beyond the chosen v1 categories. Start with the v1 terminal-effect categories and expand later if needed.

**Depends on:** M3.1, M3.3, M3.3a, M2 (the phone is where the approval UI lives).

---

## M3.5 — chat UI on the phone and the AI loop end to end

**Goal:** the Android app has a chat UI that can list chats, view messages, send a message, see streaming AI responses, see tool-request/approval prompts for gated terminal-effect tools, and answer them — completing the loop end to end.

**Acceptance criteria:**

- [ ] The app can list chats and open one.
- [ ] The app can show messages in a chat, including streaming AI responses as deltas arrive.
- [ ] The app can send a user message into a chat, which triggers the agent's LLM client + tool loop.
- [ ] The app can display a pending approval prompt for gated terminal-effect tools (run command, AI terminal writes, and any other AI-reachable state-changing operations exposed to the AI) and allow/deny it.
- [ ] The app can see the AI's final answer and any tool results in the chat.
- [ ] The end-to-end loop works against a real tmux session: user sends a message, AI can read the screen and, if allowed, run a command and read its output, and the user sees it on the phone.

**Test cases:**

- [ ] Manual end-to-end on device/emulator against the agent on LAN: create a chat, send a message that asks to read the screen, see the screen contents in the answer; send a message that asks to run a command, approve it, see the output.
- [ ] Manual verification for the streaming path: deltas arrive and the UI updates as they arrive.
- [ ] Manual verification for the Ask path: the approval prompt appears for a gated tool, allow runs it, deny blocks it.
- [ ] Manual verification for the bypass-safety expectation: an AI request that tries to affect the terminal through a gated write tool without approval is blocked/asked, not silently executed.
- [ ] Unit/UI tests for the chat UI pieces where practical (message list rendering, streaming update, approval prompt).
- [ ] Negative verification: sending to a nonexistent chat, or approving a nonexistent request, is handled cleanly.

**Out of scope:**

- [ ] A polished chat UX. v1 only needs the loop to be verifiable.
- [ ] Multiple model selection UI, chat settings UI beyond the v1 terminal-effect permission categories, etc. Those can come later.

**Depends on:** M3.1–M3.4, M3.5b, M2.

---

## M3.7 — M3 acceptance test suite and release criteria

---

## M3.5b — approval integrity: immutable requests, expiry, cancellation, atomic resolution, restart safety

**Goal:** define how pending approvals behave across restart, retry, and pane changes, so approvals cannot be duplicated or replayed unintentionally.

**Acceptance criteria:**

- [ ] Each pending approval request is bound to **immutable arguments and target identity** (the exact tool, its arguments, and the session/pane it targets). The approval is for that specific request, not a reusable capability.
- [ ] Pending requests have explicit **expiry** and **cancellation** rules. A request that is not answered within the expiry window is resolved as expired (blocked), not left hanging.
- [ ] Approval resolution is **atomic** for a given request identity: answering the same request more than once does not run the tool again. Duplicate approvals of the same request do not cause duplicate execution.
- [ ] If the agent restarts while a request is pending, pending requests do not automatically replay. A restart does not execute commands that were waiting for approval before the restart. The post-restart state is explicit (pending requests are gone/expired/queued for re-approval as a new request, depending on the chosen rule), not a silent replay.
- [ ] If the target pane/session changes or disappears while a request is pending, the request is resolved as no longer applicable (blocked/cancelled), not executed against the wrong target.

**Test cases:**

- [ ] **Duplicate-approval test:** approve the same pending request twice; the tool runs once and the second approval is a no-op or explicit duplicate error, not a second execution.
- [ ] **Restart-during-execution test:** while a command is running or a request is pending, restart the agent; assert no undelried command is auto-replayed and the post-restart state is explicit.
- [ ] **Expiry test:** a pending request that is not answered within the expiry window is resolved as expired/blocked.
- [ ] **Cancellation test:** a pending request can be cancelled explicitly, and cancellation prevents execution.
- [ ] **Target-change test:** if the target pane/session disappears while a request is pending, the request is resolved as no longer applicable, not executed against the wrong target.

**Out of scope:**

- [ ] A safety classifier. This task is about approval integrity, not content classification.

**Depends on:** M3.4.

---

## M3.6 — chat transport contract for the phone

**Goal:** define how the Android app receives chat events from the agent: the transport, endpoint shape, event identity, replay, reconnect deduplication, and cancellation.

**Acceptance criteria:**

- [ ] The plan chooses the v1 chat transport: SSE or WebSocket (or another explicit choice), and records why.
- [ ] The transport has an explicit endpoint/connection shape the Android client can use.
- [ ] Events have explicit **event IDs** and a **replay cursor** so the client can ask for missed events after a reconnect and not re-receive the same event twice.
- [ ] **Reconnect deduplication** is defined: after a reconnect, the client replays from the last acknowledged event ID, and the server does not re-emit events the client already acknowledged.
- [ ] **Cancellation** is defined: the client can end the subscription, and the server stops sending events for that subscription.
- [ ] The transport supports the broker's subscription model from M3.1 (subscribe to a chat, receive existing + new messages).

**Test cases:**

- [ ] Unit tests for event ID / replay cursor logic: replay from a cursor returns the correct missed events and does not duplicate already-acknowledged events.
- [ ] Reconnect tests: simulate a disconnect and reconnect, assert the client receives missed events and does not receive duplicates.
- [ ] Cancellation tests: ending a subscription stops events for that subscription.
- [ ] Negative tests: reconnecting with an invalid/expired cursor is handled explicitly.

**Out of scope:**

- [ ] A relay transport. v1 chat transport is LAN, directly between the phone and the agent. Relay is M4.

**Depends on:** M3.1.

---

## M3.7 — M3 acceptance test suite and release criteria

**Goal:** a focused verification suite for M3 and explicit release criteria.

**Acceptance criteria:**

- [ ] The chat broker, persistence, LLM client, tool dispatch, and permission gating all have unit tests that pass.
- [ ] There is at least one integration/contract verification path for the OpenRouter loop (with a real API key where possible) and for the tool loop against real tmux.
- [ ] The end-to-end phone loop is verified manually: chat → message → AI reads screen / runs approved command → user sees result.
- [ ] The M3 release criteria are met: the phone can have a chat with an AI that can read the terminal and, with approval, run commands on the real tmux session.

**Test cases:**

- [ ] Run the M3 unit test suite; it passes.
- [ ] Run the M1 and M2 suites again to ensure M3 did not regress the substrate; they still pass.
- [ ] Manual end-to-end on device/emulator against the agent on LAN; the full chat+AI loop works.

**Out of scope:**

- [ ] Relay, images, local model, full iTerm2 feature parity.

**Depends on:** M3.1–M3.6.

---

## M3 done-ness summary

M3 is done when:

- the agent has a persistent chat broker with subscriptions and replay;
- the agent has an OpenRouter LLM client with streaming and a function-call loop;
- the agent has a v1 AI tool surface (execute_command, insert_text_at_cursor, get_screen_contents, plus a few session read tools) that maps to real tmux operations, with permissions defined around terminal effects, not only run-command;
- terminal-affecting AI tools are gated by default (run command, AI terminal writes, and any other AI-reachable state-changing operations), with the approval flow surfaced on the phone and bypass attempts tested;
- execute_command has explicit completion semantics (command-boundary tracked where supported, otherwise "input submitted" + snapshot or a bounded wait), with slow-command, delayed-output, busy-pane, and timeout tests;
- pending approvals are immutable, have expiry/cancellation/atomic resolution, and do not auto-replay on restart or execute against a changed target;
- the phone chat UI can send messages, see streaming AI responses, see pending approval prompts for gated terminal-effect tools, and answer them;
- the full loop is verified end to end against a real tmux session on LAN;
- the chat transport contract (event IDs, replay cursor, reconnect deduplication, cancellation) is verified;
- approval integrity (no duplicate execution, no auto-replay on restart, expiry/cancellation, target-change safety) is verified;
- command completion semantics are verified (no false completion claims; slow-command/delayed-output/busy-pane/timeout behavior matches the documented contract).

When M3 is done, you have the core of the iTerm2-style AI Chat loop on Linux + tmux + Android: the AI can read the terminal and, with your approval, run commands on it from the phone, with explicit failure behavior for the hard cases (busy panes, slow commands, reconnect, duplicate approvals, restart, bypass attempts). That is the substantive product for v1.
