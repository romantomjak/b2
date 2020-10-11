package command

import (
	"github.com/mitchellh/cli"
)

type VersionCommand struct {
	Ui      cli.Ui
	Version string
}

func (c *VersionCommand) Help() string {
	return ""
}

func (c *VersionCommand) Name() string { return "version" }

func (c *VersionCommand) Run(_ []string) int {
	c.Ui.Output(c.Version)
	return 0
}

func (c *VersionCommand) Synopsis() string {
	return "Prints the client version"
}
