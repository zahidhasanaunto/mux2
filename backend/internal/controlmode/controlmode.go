package controlmode

// ControlMode is the real tmux control-mode client package for M1 session/window/pane
// ownership. The scaffold currently uses a CLI-based stub in internal/tmux for the
// M1.2 probe; this package is the target implementation.
//
// The contract this package must satisfy, from the plan:
//
// - create a detached session with a name + start command and return a stable
//   session identifier that maps to the real tmux session;
// - list sessions the agent owns or can see, with session id, name, state,
//   creation time where available;
// - kill a session by its stable identifier;
// - on agent startup, re-resolve existing tmux sessions to the agent's
//   identifiers so session identifiers are stable across restarts.
//
// The session↔window relationship is modeled as links, not strict ownership:
// a window can be linked into more than one session. Per the tmux 3.6 probe on
// this host, shared windows keep a single window identity across sessions and
// each session has its own active pane.
//
// M1.7 adds HTTP around these operations; M1.8 adds the GUI that calls the same
// API. This package does not own auth or transport; it is the tmux integration
// layer the agent calls.

// TODO: implement the control-mode client.
//   - use `tmux -L <socket> -f /dev/null control-mode` or the equivalent
//     long-lived control socket, depending on what tmux 3.6 supports cleanly;
//   - parse the control stream for session/window/pane lifecycle events;
//   - map tmux session ids to stable internal session identifiers;
//   - expose CreateSession, ListSessions, KillSession, and a startup
//     re-resolution step.
//   - keep the CLI stub in internal/tmux only as a probe helper, not the
//     production ownership path, once this package is real.

const todo = "control-mode client not yet implemented"
