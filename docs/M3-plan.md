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

## M3.3 — AI tool surface for v1

**Goal:** a set of LLM 도구 (tool/function) definitions that map LLM tool calls to the agent's tmux operations, starting with the iTerm2-style v1 surface.

**Acceptance criteria:**

- [ ] The tool surface includes at minimum: `execute_command`, `insert_text_at_cursor`, `get_screen_contents`, plus a few session-state/read tools such as `get_terminal_size`, `get_current_directory`, and `get_last_command` where the underlying tmux/agent semantics allow it.
- [ ] Each tool has a clear name, description, and parameter schema, and is registered with the LLM client.
- [ ] `execute_command` maps to sending a command into the selected pane and capturing its output via the M1 send+capture flow.
- [ ] `insert_text_at_cursor` maps to sending text into the selected pane as keystrokes through the M1 send-keys flow.
- [ ] `get_screen_contents` maps to capturing the pane's visible screen (and scrollback where relevant) through the M1 capture flow.
- [ ] The tool surface can be targeted at a specific session/pane (the AI's "current session"), resolved from the chat's linked session.
- [ ] Tool results are formatted so the LLM can use them (clear output, bounded size where relevant).

**Test cases:**

- [ ] Unit tests for each tool's argument parsing and validation.
- [ ] Unit tests for the tool dispatch logic: given a tool call, the dispatcher invokes the correct agent operation and returns a result.
- [ ] Integration tests against a real tmux-backed agent for the key tools: execute_command produces output, insert_text_at_cursor types text, get_screen_contents returns the screen.
- [ ] Negative tests: tool calls with missing/invalid arguments return clear errors; tools targeted at a dead/nonexistent session return clear errors.

**Out of scope:**

- [ ] The full 24-case iTerm2 remote-command catalog (man pages, file creation, clipboard, restart session, browser actions, SSH detection, etc.). Those are M5 if we want them.
- [ ] Shell-integration marks. tmux does not provide those natively; if we want output-boundary handling better than "send + capture", that's a design decision for M3/M5, not assumed here.

**Depends on:** M3.1, M3.2, M1.4, M1.5, M2.2.

---

## M3.4 — permission gating for execute_command

**Goal:** execute_command is gated by an Ask/Allow/Deny model, with the approval flow surfaced on the phone, so the AI cannot run commands on the machine without consent in v1.

**Acceptance criteria:**

- [ ] Each chat has a permission setting for execute_command and possibly other dangerous tools: Allow, Ask, Deny (and, if you want, an "Always Allow" path that is still safety-gated later).
- [ ] When the LLM requests execute_command and the chat setting is Ask, the request is parked and the phone is asked to approve or deny before the command runs.
- [ ] When allowed, the command runs through M3.3's execute_command tool and the result is fed back into the chat.
- [ ] When denied, the tool result reflects the denial and the chat continues.
- [ ] The approval/dec denial is addressable by chat and message so the phone can answer the specific pending request.
- [ ] The permission state is persisted (so a setting chosen on the phone survives restarts).

**Test cases:**

- [ ] Unit tests for the permission state machine: Allow runs immediately, Ask parks and waits, Deny blocks, and the transitions are correct.
- [ ] Unit tests for the pending-request resolution: answering a specific pending request runs or blocks the right command.
- [ ] Integration tests against the real agent + tmux for the Allow path: an approved execute_command runs and returns output.
- [ ] Negative tests: answering a nonexistent or already-resolved request is handled cleanly.

**Out of scope:**

- [ ] A safety classifier or "Always Allow still blocked by safety" layer. That is a later enhancement if you want it; v1 Ask gating is the core safety behavior.
- [ ] Per-category permissions beyond what you want for v1. Start with execute_command gating and expand later if needed.

**Depends on:** M3.1, M3.3, M2 (the phone is where the approval UI lives).

---

## M3.5 — chat UI on the phone and the AI loop end to end

**Goal:** the Android app has a chat UI that can list chats, view messages, send a message, see streaming AI responses, see tool-request/approval prompts for execute_command, and answer them — completing the loop end to end.

**Acceptance criteria:**

- [ ] The app can list chats and open one.
- [ ] The app can show messages in a chat, including streaming AI responses as deltas arrive.
- [ ] The app can send a user message into a chat, which triggers the agent's LLM client + tool loop.
- [ ] The app can display a pending execute_command approval prompt and allow/deny it.
- [ ] The app can see the AI's final answer and any tool results in the chat.
- [ ] The end-to-end loop works against a real tmux session: user sends a message, AI can read the screen and, if allowed, run a command and read its output, and the user sees it on the phone.

**Test cases:**

- [ ] Manual end-to-end on device/emulator against the agent on LAN: create a chat, send a message that asks to read the screen, see the screen contents in the answer; send a message that asks to run a command, approve it, see the output.
- [ ] Manual verification for the streaming path: deltas arrive and the UI updates as they arrive.
- [ ] Manual verification for the Ask path: the approval prompt appears, allow runs the command, deny blocks it.
- [ ] Unit/UI tests for the chat UI pieces where practical (message list rendering, streaming update, approval prompt).
- [ ] Negative verification: sending to a nonexistent chat, or approving a nonexistent request, is handled cleanly.

**Out of scope:**

- [ ] A polished chat UX. v1 only needs the loop to be verifiable.
- [ ] Multiple model selection UI, chat settings UI beyond the execute_command permission, etc. Those can come later.

**Depends on:** M3.1–M3.4, M2.

---

## M3.6 — M3 acceptance test suite and release criteria

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

**Depends on:** M3.1–M3.5.

---

## M3 done-ness summary

M3 is done when:

- the agent has a persistent chat broker with subscriptions and replay;
- the agent has an OpenRouter LLM client with streaming and a function-call loop;
- the agent has a v1 AI tool surface (execute_command, insert_text_at_cursor, get_screen_contents, plus a few session read tools) that maps to real tmux operations;
- execute_command is gated by an Ask/Allow/Deny flow surfaced on the phone, with persistence;
- the phone chat UI can send messages, see streaming AI responses, see pending execute_command approvals, and answer them;
- the full loop is verified end to end against a real tmux session on LAN.

When M3 is done, you have the core of the iTerm2-style AI Chat loop on Linux + tmux + Android: the AI can read the terminal and, with your approval, run commands on it from the phone. That is the substantive product for v1.
