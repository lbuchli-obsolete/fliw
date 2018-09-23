package cli

import (
	"fmt"
	"regexp"
)

const MainHelp = `
FLIW - Framework for Low-Interaction Windows

The FLIW command-line interface supports the following commands:

run   - runs a fliw package[1]. See "fliw help run"
build - builds a fliw package[1]. See "fliw help build"
test  - tests a fliw package[1] for errors and reports them. See "fliw help test"

For dependencies, see "fliw help deps"

[1] fliw package
A fliw package is a package runnable by fliw. It is a folder with a given file structure.
For more information on how to make a fliw package, please visit the fliw github wiki.
`

const RunHelp = `
WIP
`

const BuildHelp = `
WIP
`

const TestHelp = `
WIP
`

const DepsHelp = `
WIP
`

var validAbsoluteDirPathRegex *regexp.Regexp

// initialize.
func init() {
	validAbsoluteDirPathRegex = regexp.MustCompile(`/([[:graph:]]+/?)*`)
}

// RunCommand runs a command given the subcommand and its arguments.
// If the command is unknown it displays a help instead.
func RunCommand(command ...string) {
	if len(command) > 1 {
		switch command[0] {
		case "run":
			run(command[1:])
		case "build":
			build(command[1:])
		case "test":
			test(command[1:])
		case "help":
			displayHelp(command[1])
		}
	} else {
		displayHelp()
	}
}

// displayHelp displays help in form of text.
func displayHelp(subcommand ...string) {
	switch subcommand[0] {
	case "":
		fmt.Println(MainHelp)
	case "run":
		fmt.Println(RunHelp)
	case "build":
		fmt.Println(BuildHelp)
	case "test":
		fmt.Println(TestHelp)
	case "deps":
		fmt.Println(DepsHelp)
	default:
		fmt.Println(MainHelp)
	}
}
