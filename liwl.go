/*
Low Interaction Window Launcher
This is the main executable. It allows you to launch your app
You'll just have to give it the path to the root folder of your app
*/

package main

import (
	"fmt"
	"log"
	"os"
	"plugin"

	"github.com/phoenixdevelops/fliw/data"
	"github.com/phoenixdevelops/fliw/parser"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

func main() {
	args := os.Args[1:]
	log.Println(args)
	// TODO launch app

	// Initialization
	err := initialize()
	if err != nil {
		log.Fatal(err)
	}

	cont, bgcolor, windowtype, err := parser.ParseXMLFile("/home/lukas/go/src/fliwexamples/basic/style.xml")
	fmt.Println("Container: ", cont)
	fmt.Println("Color: ", bgcolor)
	fmt.Println("Windowtype", windowtype)
	fmt.Println("Error: ", err)

	plug, err := plugin.Open("/home/lukas/go/src/fliwexamples/basic/app.so")
	if err != nil {
		log.Fatal(err)
		return
	}

	data.ShowWindow(*cont, bgcolor, *plug)
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
