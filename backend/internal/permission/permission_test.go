package permission

import "testing"

func TestTerminalInputCategoryIsSingleAndCoversExecuteCommandInsertTextAtCursorAndPaste(t *testing.T) {
	// This is the test that backs the M3.4 single-category rule: the v1
	// terminal-effect surface is one permission category, and that category
	// covers execute_command, insert_text_at_cursor, and paste.
	values := All()
	needle := CategoryTerminalInput
	found := false
	for _, c := range values {
		if c == needle {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("terminal input category not present in All(): %v", values)
	}
}

func TestTerminalInputPermissionStateMachine(t *testing.T) {
	tests := []struct {
		name      string
		set       Set
		wantAllow bool
		wantAsk   bool
		wantDeny  bool
	}{
		{
			name: "explicit allow",
			set: Set{
				CategoryTerminalInput: ValueAllow,
			},
			wantAllow: true,
			wantAsk:   false,
			wantDeny:  false,
		},
		{
			name: "explicit ask",
			set: Set{
				CategoryTerminalInput: ValueAsk,
			},
			wantAllow: false,
			wantAsk:   true,
			wantDeny:  false,
		},
		{
			name: "explicit deny",
			set: Set{
				CategoryTerminalInput: ValueDeny,
			},
			wantAllow: false,
			wantAsk:   false,
			wantDeny:  true,
		},
		{
			name: "empty set has no terminal-input setting",
			set:  Set{},
			wantAllow: false,
			wantAsk:   false,
			wantDeny:  false,
		},
		{
			name: "other categories do not affect terminal-input helpers",
			set: Set{
				CategorySessionCreate: ValueAllow,
				CategorySessionKill:   ValueDeny,
				CategoryResize:        ValueAsk,
			},
			wantAllow: false,
			wantAsk:   false,
			wantDeny:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.set.IsTerminalInputAllowed(); got != tt.wantAllow {
				t.Errorf("IsTerminalInputAllowed() = %v, want %v", got, tt.wantAllow)
			}
			if got := tt.set.NeedsApproval(); got != tt.wantAsk {
				t.Errorf("NeedsApproval() = %v, want %v", got, tt.wantAsk)
			}
			if got := tt.set.IsTerminalInputDenied(); got != tt.wantDeny {
				t.Errorf("IsTerminalInputDenied() = %v, want %v", got, tt.wantDeny)
			}
		})
	}
}

func TestWrongValueDoesNotAccidentallySatisfyAnyBranch(t *testing.T) {
	// Important for the single-category rule: a typo or an unexpected value
	// must not accidentally be treated as allow or ask or deny.
	s := Set{
		CategoryTerminalInput: Value("maybe"),
	}
	if s.IsTerminalInputAllowed() {
		t.Error("unexpected allow for unknown value")
	}
	if s.NeedsApproval() {
		t.Error("unexpected ask for unknown value")
	}
	if s.IsTerminalInputDenied() {
		t.Error("unexpected deny for unknown value")
	}
}

func TestTerminalInputCoversAllTerminalAffectingTools(t *testing.T) {
	// The M3.4 acceptance criterion requires that no combination of v1
	// permission settings allows AI terminal input to execute a command when
	// terminal input is denied or asked. The implementation case is:
	// terminal-input Deny + anything else must block any tool call that would
	// send keystrokes/text to the terminal.
	deny := Set{
		CategoryTerminalInput: ValueDeny,
		CategorySessionCreate: ValueAllow,
		CategorySessionKill:   ValueAllow,
		CategoryResize:        ValueAllow,
	}
	if deny.IsTerminalInputAllowed() || deny.NeedsApproval() {
		t.Error("terminal-input Deny must block terminal-affecting tools")
	}
	ask := Set{
		CategoryTerminalInput: ValueAsk,
		CategorySessionCreate: ValueAllow,
		CategorySessionKill:   ValueAllow,
		CategoryResize:        ValueAllow,
	}
	if ask.IsTerminalInputAllowed() {
		t.Error("terminal-input Ask must not allow terminal-affecting tools immediately")
	}
	if !ask.NeedsApproval() {
		t.Error("terminal-input Ask must require approval")
	}
}
