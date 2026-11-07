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
- [ ] **Durable vs. best-effort events are distinguished.** For v1:
  - **Durable:** chat messages (user messages, AI answers, tool results, approval-request and approval-resolution events) are persisted and can be replayed after restart. The replay cursor in M3.6 is defined against these durable event identities.
  - **Best-effort / reconstruction:** streaming deltas for a live AI answer are not necessarily individually durable in v1. If the agent restarts mid-stream, the client's replay after reconnect is defined as: the durable final answer (or the durable approval/tool-result events) can be replayed, but individual streaming deltas that were not durably committed may be lost and are reconstructed as a fresh delivery of the durable result where applicable, not as a perfect byte-identical replay of every delta.
  - **Ack-lost deduplication:** if a client received an event but its acknowledgment was lost (for example due to a disconnect), the server may re-emit the event on reconnect. The client must deduplicate by event ID so a re-emitted event does not appear twice. This is part of the M3.6 reconnect/deduplication contract.
- [ ] After an agent restart, the replay cursor behavior is explicit: the client replays durable events from the last acknowledged durable event ID, and any best-effort deltas that were in flight are handled per the reconstruction rule above, not silently dropped or duplicated.

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
- [ ] The permission model is defined around **AI terminal input as one category**, not as independent "run command" vs. "AI terminal writes" categories that can be set differently and create a bypass.
  - In v1, there is a single terminal-input permission category covering all AI-supplied input that can reach the terminal as keystrokes/text, including `execute_command`, `insert_text_at_cursor`, and paste paths that can submit or carry control characters.
  - If a chat's terminal-input setting is Deny, the AI cannot send any terminal input through the tool surface. If it is Ask, any such tool call is parked for approval. If it is Allow, the AI may use the allowed input tools.
  - This single-category rule exists specifically so a configuration like "run command Deny + terminal writes Allow" cannot allow `insert_text_at_cursor` to submit a command as a bypass. There is no valid v1 configuration in which one category is Deny and another is Allow in a way that lets AI terminal input execute commands.
  - If finer granularity is wanted later, it must be added as explicit precedence rules with their own tests, not as independent categories that can conflict.
  - Session create/kill and resize are separate state-changing operations and are gated where they are reachable through the AI tool surface, but they are not a substitute path for terminal input.
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
- [ ] **Bypass tests for the single-category rule:** verify that no combination of v1 permission settings allows AI terminal input to execute a command when terminal input is denied or asked.
  - Primary case: terminal-input Deny + anything else. Any AI tool call that would send keystrokes/text to the terminal (`execute_command`, `insert_text_at_cursor`, paste with submission/control characters) is blocked or asked, not executed.
  - Conflicting-configuration case: there is no valid v1 state in which one category is Deny and another is Allow in a way that lets AI terminal input run a command. The implementation must reject or ignore any such inconsistent configuration, and the test asserts that inconsistent configs cannot produce execution.
  - Smuggling case: payloads that try to submit a command through `insert_text_at_cursor` or paste (for example a command string followed by a submission character) are treated as terminal input under the same single category and are blocked/asked when terminal input is Deny/Ask.
  - Allow case: when terminal-input is Allow, an approved run-command tool runs and returns output under the M3.3a completion semantics.
- [ ] Negative tests: answering a nonexistent or already-resolved request is handled cleanly.

**Out of scope:**

- [ ] A safety classifier or "Always Allow still blocked by safety" layer. That is a later enhancement if you want it; v1 Ask gating is the core safety behavior.
- [ ] Per-tool permissions beyond the chosen v1 terminal-input category and the separate state-changing operations (session create/kill, resize) that are gated where reachable.

**Depends on:** M3.1, M3.3, M3.3a, M2 (the phone is where the approval UI lives).

---

## M3.5 — chat UI on the phone and the AI loop end to end

**Goal:** the Android app has a chat UI that can list chats, create a chat with a target pane, view messages, send a message, see streaming AI responses, see tool-request/approval prompts for gated terminal-effect tools, answer/cancel them, and cancel an in-progress run — completing the loop end to end. This task depends on the chat transport contract in M3.6 (both the subscription side and the write side).

**Acceptance criteria:**

- [ ] The app can list chats and open one (using the M3.6 write API).
- [ ] The app can create a chat with an optional target session/pane for the AI (using the M3.6 write API), so the AI's "current session" is explicit.
- [ ] The app can show messages in a chat, including streaming AI responses as deltas arrive (using the M3.6 subscription transport).
- [ ] The app can send a user message into a chat (using the M3.6 write API), which triggers the agent's LLM client + tool loop.
- [ ] The app can display a pending approval prompt for gated terminal-effect tools (the single terminal-input category, plus any other AI-reachable state-changing operations exposed to the AI) and allow/deny/cancel it (using the M3.6 approval-resolution API).
- [ ] The app can cancel an in-progress AI run (using the run-cancellation contract from M3.5b), and the cancellation is reflected in the chat.
- [ ] The app can see the AI's final answer and any tool results in the chat.
- [ ] The end-to-end loop works against a real tmux session: user sends a message, AI can read the screen and, if allowed, run a command and read its output, and the user sees it on the phone.

**Test cases:**

- [ ] Manual end-to-end on device/emulator against the agent on LAN: create a chat with a target pane, send a message that asks to read the screen, see the screen contents in the answer; send a message that asks to run a command, approve it, see the output.
- [ ] Manual verification for the streaming path: deltas arrive and the UI updates as they arrive.
- [ ] Manual verification for the Ask path: the approval prompt appears for a gated tool, allow runs it, deny blocks it.
- [ ] Manual verification for the bypass-safety expectation: an AI request that tries to affect the terminal through a gated write tool without approval is blocked/asked, not silently executed.
- [ ] Manual verification for run cancellation: cancel an in-progress run and observe that no further tool calls execute and the chat reflects the cancellation.
- [ ] Unit/UI tests for the chat UI pieces where practical (message list rendering, streaming update, approval prompt, run cancellation).
- [ ] Negative verification: sending to a nonexistent chat, creating a chat with a bad target pane, or approving/resolving a nonexistent request, is handled cleanly.

**Out of scope:**

- [ ] A polished chat UX. v1 only needs the loop to be verifiable.
- [ ] Multiple model selection UI, chat settings UI beyond the v1 terminal-input category and the separate state-changing operations that are gated where reachable.
- [ ] Relay, images, local model, full iTerm2 feature parity.

**Depends on:** M3.1–M3.4, M3.5b, M3.6, M2.---



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

## M3.6 — chat transport contract for the phone: subscriptions and writes

**Goal:** define both how the Android app receives chat events and how it writes to the chat: connection shape, event identity, replay, reconnect deduplication, cancellation, and the chat write API (create chat with a target pane, submit messages, resolve approvals), plus retry behavior.

**Acceptance criteria:**

- [ ] The plan chooses the v1 chat transport: SSE or WebSocket (or another explicit choice), and records why.
- [ ] The transport has an explicit endpoint/connection shape the Android client can use.
- [ ] **Subscription side:** events have explicit **event IDs** and a **replay cursor**, so the client can reconnect and ask for missed events without re-receiving the same event twice. Reconnect deduplication is defined: after a reconnect, the client replays from the last acknowledged event ID, and the server does not re-emit events the client already acknowledged. Cancellation is defined: the client can end the subscription and the server stops sending events for it.
- [ ] **Write side:** there are explicit request/response contracts for:
  - creating a chat (optionally with a target session/pane the AI can be linked to),
  - submitting a user message to a chat,
  - resolving a pending approval (allow/deny/cancel) for a specific pending request identity,
  - listing chats, listing messages in a chat, and reading the current pending requests for a chat.
- [ ] The write API has explicit retry behavior: idempotent operations (where applicable, such as resolving a specific pending request identity) can be retried safely; non-idempotent submissions have a defined behavior on retry (for example, do not silently duplicate a user message; return the existing result or an explicit conflict).
- [ ] The transport supports the broker's subscription model from M3.1 (subscribe to a chat, receive existing + new messages) and the write operations above.
- [ ] Chat transport connections require the same auth as the rest of the agent API (M1.7/M1.8): no open chat transport.

**Test cases:**

- [ ] Unit tests for event ID / replay cursor logic: replay from a cursor returns the correct missed events and does not duplicate already-acknowledged events.
- [ ] Reconnect tests: simulate a disconnect and reconnect, assert the client receives missed events and does not receive duplicates.
- [ ] Cancellation tests: ending a subscription stops events for that subscription.
- [ ] Negative tests: reconnecting with an invalid/expired cursor is handled explicitly.
- [ ] Write-API contract tests: create chat, submit message, and resolve approval requests produce the documented response shapes and error shapes.
- [ ] Retry tests: retrying an idempotent approval resolution is safe; retrying a message submission does not silently duplicate the message.
- [ ] Auth tests for the chat transport: connections and write requests without valid credentials are rejected.

**Out of scope:**

- [ ] A relay transport. v1 chat transport is LAN, directly between the phone and the agent. Relay is M4.

**Depends on:** M3.1, M1.7, M1.8.

---

## M3.7 — M3 acceptance test suite and release criteria

**Goal:** a focused verification suite for M3 and explicit release criteria.

**Acceptance criteria:**

- [ ] The chat broker, persistence, chat transport (subscriptions + writes), LLM client, tool dispatch, terminal-input permission gating, command-completion semantics, approval integrity, run cancellation, iteration/time limits, and durable-event guarantees all have unit/contract tests that pass.
- [ ] There is at least one integration/contract verification path for the OpenRouter loop (with a real API key where possible), for the tool loop against real tmux, and for the chat transport reconnect/replay/retry behavior.
- [ ] The end-to-end phone loop is verified manually: create a chat with a target pane → send a message → AI reads the screen and, with approval, runs a command → user sees the result; plus run cancellation and bypass-safety behavior.
- [ ] The M3 release criteria are met: the phone can have a chat with an AI that can read the terminal and, with approval, run commands on the real tmux session, with the full set of failure/integrity contracts above.

**Test cases:**

- [ ] Run the M3 unit/contract test suite; it passes.
- [ ] Run the M1 and M2 suites again to ensure M3 did not regress the substrate; they still pass.
- [ ] Manual end-to-end on device/emulator against the agent on LAN; the full chat+AI loop works, including cancellation and bypass-safety expectations.

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
- the phone chat UI can send messages, see streaming AI responses, see pending approval prompts for gated terminal-effect tools, answer/cancel them, and cancel an in-progress run; the run-cancellation, iteration-limit, and time-limit contracts are verified;
- the full loop is verified end to end against a real tmux session on LAN;
- the chat transport contract (event IDs, replay cursor, reconnect deduplication, cancellation, write API, retry behavior, auth) is verified;
- approval integrity (no duplicate execution, no auto-replay on restart, expiry/cancellation, target-change safety) is verified;
- command completion semantics are verified (no false completion claims; slow-command/delayed-output/busy-pane/timeout behavior matches the documented contract);
- durable vs. best-effort event guarantees are verified (chat messages and approval events survive restart and can be replayed; streaming deltas may be best-effort with a defined reconstruction/replay behavior; cursor behavior after restart and ack-lost deduplication are defined and tested).

When M3 is done, you have the core of the iTerm2-style AI Chat loop on Linux + tmux + Android: the AI can read the terminal and, with your approval, run commands on it from the phone, with explicit failure behavior for the hard cases (busy panes, slow commands, reconnect, duplicate approvals, restart, bypass attempts, runaway loops, run cancellation, event durability). That is the substantive product for v1.
