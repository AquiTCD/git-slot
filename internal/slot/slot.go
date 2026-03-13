package slot

type SlotState int

const (
	SlotEmpty SlotState = iota
	SlotActive
)

type Slot struct {
	Name     string
	Icon     string
	Path     string
	State    SlotState
	Branch   string
	IsDirty  bool
	HeadHash string
}

func (s *Slot) DisplayState() string {
	if s.State == SlotEmpty {
		return "empty"
	}
	if s.IsDirty {
		return "dirty"
	}
	return "active"
}
