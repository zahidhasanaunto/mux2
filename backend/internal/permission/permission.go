package permission

// Permission is the v1 terminal-effect permission model.
//
// M1 has no chat/LLM/approval layer yet. This package is the early shape of the
// permission model so M3.4 / M3.5 can plug into it. The interesting rule for v1
// is the single terminal-input category: execute_command, insert_text_at_cursor,
// and paste are one permission category, not independent conflicting ones.
// Any later AI-reachable resize or session create/kill operations are separate
// permissions (when exposed).
//
// v1 permission values are one of Allow, Ask, Deny for the terminal-input
// category. M3 adds the Ask → parked-for-approval flow and the run-cancellation /
// iteration/time-limit rules.

type Category int

const (
	CategoryTerminalInput Category = 1 << iota
	CategorySessionCreate
	CategorySessionKill
	CategoryResize
)

// All returns all defined categories.
func All() []Category {
	return []Category{
		CategoryTerminalInput,
		CategorySessionCreate,
		CategorySessionKill,
		CategoryResize,
	}
}

// Value is one of Allow, Ask, Deny.
type Value string

const (
	ValueAllow Value = "allow"
	ValueAsk   Value = "ask"
	ValueDeny  Value = "deny"
)

// Set is the per-chat permission state for the categories that exist in v1.
type Set map[Category]Value

// TerminalInput returns the single terminal-input category value.
func (s Set) TerminalInput() Value {
	if v, ok := s[CategoryTerminalInput]; ok {
		return v
	}
	return ""
}

// IsTerminalInputAllowed returns true when the terminal-input category allows
// terminal-affecting tool execution without an approval flow.
func (s Set) IsTerminalInputAllowed() bool {
	return s.TerminalInput() == ValueAllow
}

// IsTerminalInputDenied returns true when the terminal-input category blocks
// terminal-affecting tool execution.
func (s Set) IsTerminalInputDenied() bool {
	return s.TerminalInput() == ValueDeny
}

// NeedsApproval returns true when the terminal-input category requires an
// approval flow before terminal-affecting tool execution.
func (s Set) NeedsApproval() bool {
	return s.TerminalInput() == ValueAsk
}

// Unit tests for the permission state machine live in permission_test.go.
// Those tests assert Allow runs immediately, Ask parks and waits, Deny blocks,
// and the transitions are correct for the single terminal-input category.
