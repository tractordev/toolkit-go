package event

type Type int

const (
	None Type = iota
	// window
	Close
	Created
	Destroyed
	Focused
	Blurred
	Resized
	Moved
	// menu
	MenuItem
	// shell
	Shortcut
)

func (e Type) String() string {
	return []string{
		"",
		"close",
		"create",
		"destroy",
		"focus",
		"blur",
		"resize",
		"move",
		"menu",
		"shortcut",
	}[e]
}
