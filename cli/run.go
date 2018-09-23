package cli

import (
	"log"
	"runtime/debug"

	"github.com/phoenixdevelops/fliw/launcher"
)

func run(args []string) {
	if validAbsoluteDirPathRegex.MatchString(args[0]) {
		err := launcher.ShowWindow(args[0])
		if err != nil {
			debug.PrintStack()
			log.Fatal(err)
		}
	} else {
		log.Fatal("Specified path does not lead to a directory.")
	}
}
