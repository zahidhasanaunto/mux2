package tmux

import (
	"os"
	"os/exec"
	"strings"
	"time"
)

// ControlModeSessionOwnershipPlanTest is an integration-only probe used by the
// M1.2 plan tests to prove the agent can own tmux sessions through control mode.
//
// This is intentionally a thin wrapper around the tmux CLI for the scaffold pass.
// The real implementation is a control-mode client that holds a long-lived control
// socket and parses the control output stream. The plan tests assert behavior,
// not this specific implementation, so we can start here and replace the internals
// with a real control-mode client once the parsing layer exists.
//
// Isolation rule for plan tests: use a dedicated tmux socket (for example
// MUX2_TMUX_SOCKET=mux2-plan-test with tmux -L mux2-plan-test -f /dev/null).
// That keeps plan tests from touching your main tmux server.

const defaultSocketEnv = "MUX2_TMUX_SOCKET"

// Isolate returns the tmux command prefix for a dedicated socket when the
// MUX2_TMUX_SOCKET environment variable is set. Otherwise it returns an empty
// prefix so commands use the default server.
func Isolate() []string {
	socket := os.Getenv(defaultSocketEnv)
	if socket == "" {
		return nil
	}
	return []string{"-L", socket, "-f", "/dev/null"}
}

// StartServer starts a tmux server on the chosen socket (or the default server
// when no isolation is configured). It is best-effort for the scaffold; real
// ownership uses control mode.
func StartServer() error {
	return run("start-server")
}

// ListSessions lists tmux sessions visible to the configured server, as
// tab-separated lines in tmux list-sessions form.
func ListSessions() (string, error) {
	return runWithOutput("list-sessions")
}

// KillServer kills the tmux server for the configured socket. Best-effort for
// the scaffold.
func KillServer() error {
	return run("kill-server")
}

func prependIsolate(args []string) []string {
	return append(Isolate(), args...)
}

func run(args ...string) error {
	cmd := exec.Command("tmux", prependIsolate(args)...)
	cmd.Stderr = nil
	return cmd.Run()
}

func runWithOutput(args ...string) (string, error) {
	cmd := exec.Command("tmux", prependIsolate(args)...)
	out, err := cmd.Output()
	return string(out), err
}

// ParseListSessions parses the output of `tmux list-sessions -F '#{session_name}\t#{session_id}\t#{session_created}\t#{session_windows}\t#{session_height}x#{session_width}'`
// into a list of raw session lines. This is the parsing hook the real
// control-mode client will replace with streamed control output parsing.
func ParseListSessions(output string) []string {
	var out []string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

// WaitForSession is a helper for plan tests that need to wait for a session to
// appear on the tmux server within a bounded time.
func WaitForSession(pred func() (bool, error), timeout time.Duration) (bool, error) {
	deadline := time.Now().Add(timeout)
	for {
		ok, err := pred()
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
		if time.Now().After(deadline) {
			return false, nil
		}
		time.Sleep(20 * time.Millisecond)
	}
}
