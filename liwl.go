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

	cont, bgcolor, windowtype, err := parser.ParseXMLFile("/home/lukas/go/src/github.com/phoenixdevelops/fliw/example/basic/style.xml")
	fmt.Println("Container: ", cont)
	fmt.Println("Color: ", bgcolor)
	fmt.Println("Windowtype", windowtype)
	fmt.Println("Error: ", err)

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
