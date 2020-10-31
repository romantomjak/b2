package command

import (
	"github.com/mitchellh/cli"
	"github.com/romantomjak/b2/version"
)

// Commands returns the mapping of CLI commands for B2
func Commands(ui cli.Ui) map[string]cli.CommandFactory {
	baseCommand := &baseCommand{
		ui: ui,
	}

	commands := map[string]cli.CommandFactory{
		"create": func() (cli.Command, error) {
			return &CreateBucketCommand{
				baseCommand: baseCommand,
			}, nil
		},
		"list": func() (cli.Command, error) {
			return &ListCommand{
				baseCommand: baseCommand,
			}, nil
		},
		"get": func() (cli.Command, error) {
			return &GetCommand{
				baseCommand: baseCommand,
			}, nil
		},
		"put": func() (cli.Command, error) {
			return &PutCommand{
				baseCommand: baseCommand,
			}, nil
		},
		"version": func() (cli.Command, error) {
			return &VersionCommand{
				baseCommand: baseCommand,
				Version:     version.FullVersion(),
			}, nil
		},
	}

	return commands
}
