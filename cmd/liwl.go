/*
Low Interaction Window Launcher
This is the main executable. It allows you to launch your app
You'll just have to give it the path to the root folder of your app
*/

package main

import (
	"log"
	"os"
)

func main() {
	args := os.Args[1:]
	log.Println(args)
	// TODO launch app
}
