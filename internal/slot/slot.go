package slot

type SlotState int

const (
	SlotEmpty SlotState = iota
	SlotActive
)

type Slot struct {
	Name        string
	Icon        string
	Path        string
	State       SlotState
	Branch      string
	HeadHash    string
	DirtyCount  int
	AheadCount  int
	BehindCount int
	HasUpstream bool
}

func (s *Slot) DisplayState() string {
	if s.State == SlotEmpty {
		return "empty"
	}
	if s.DirtyCount > 0 {
		return "dirty"
	}
	return "active"
}
