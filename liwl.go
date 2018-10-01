/*
Low Interaction Window Launcher
This is the main executable. It allows you to launch your app
You'll just have to give it the path to the root folder of your app
*/

package main

import (
	"log"
	"os"

	"github.com/phoenixdevelops/fliw/cli"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

func main() {
	args := os.Args[1:]
	log.Println(args)

	// Initialization
	err := initialize()
	if err != nil {
		log.Fatal(err)
	}

	// TODO use arg instead of predefined strings
	//cli.RunCommand("build", "/home/lukas/swayland/files/home/.scripts/sidebar/power")
	//cli.RunCommand("run", "/home/lukas/swayland/files/home/.scripts/sidebar/power")
	cli.RunCommand("build", "/home/lukas/go/src/fliwexamples/basic")
	cli.RunCommand("run", "/home/lukas/go/src/fliwexamples/basic")
}

func initialize() (err error) {
	err = sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		return err
	}

	err = ttf.Init()
	if err != nil {
		return err
	}

	return nil
}
