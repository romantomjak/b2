package command

type VersionCommand struct {
	*baseCommand
	Version string
}

func (c *VersionCommand) Help() string {
	return ""
}

func (c *VersionCommand) Name() string { return "version" }

func (c *VersionCommand) Run(_ []string) int {
	c.ui.Output(c.Version)
	return 0
}

func (c *VersionCommand) Synopsis() string {
	return "Prints the client version"
}
