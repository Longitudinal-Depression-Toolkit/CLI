package commandpalette

type Entry struct {
	Label   string
	Command []string
	Aliases []string
}

type ResolvedCommand struct {
	Args []string
	Raw  string
}
