package main

import (
	"fmt"

	client "github.com/level11consulting/ocelot/client/ui"
	"github.com/level11consulting/ocelot/version"
	"github.com/mitchellh/cli"

	"os"
)

func mainDo() int {
	args := os.Args[1:]

	// todo how to get just keys out of map?
	var cmds []string
	for c := range client.Commands {
		cmds = append(cmds, c)
	}
	clie := &cli.CLI{
		Args:         args,
		Commands:     client.Commands,
		Autocomplete: true,
		Name:         "ocelot",
		Version:      version.GetShort(),
		HelpFunc:     cli.FilteredHelpFunc(cmds, cli.BasicHelpFunc("ocelot")),
	}

	exitCode, err := clie.Run()
	if err != nil {
		fmt.Errorf("wah")
	}
	return exitCode
}

func main() {
	os.Exit(mainDo())
}
