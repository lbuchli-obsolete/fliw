/*
Low Interaction Window Launcher
This is the main executable. It allows you to launch your app
You'll just have to give it the path to the root folder of your app
*/

package main

import (
	"log"
	"os"

	"github.com/phoenixdevelops/fliw/launch"
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

	// TODO change path to args[0]
	err = launch.ShowWindow("/home/lukas/go/src/fliwexamples/basic")
	if err != nil {
		log.Fatal(err)
	}
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
