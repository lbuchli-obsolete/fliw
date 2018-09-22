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

	"github.com/phoenixdevelops/fliw/backend"
	"github.com/phoenixdevelops/fliw/data"
	"github.com/phoenixdevelops/fliw/parser"
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
	err = ShowWindow("/home/lukas/go/src/fliwexamples/basic")
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

/*
##############################################################
# Section: Launcher
##############################################################
*/

var plugin *backend.Plugin

// ShowWindow Shows a window given the path to the config files
func ShowWindow(path string) (err error) {
	// process the files
	// these steps are fatal:
	// if an error occurs, the program can't continue

	backend.AddPlugin(path + "/app.so")

	window, windowtype, err := parser.UnmarshalXMLFile(path)
	if err != nil {
		return
	}

	// Get a function for initializing, updating and eventhandling

	var initialize func(*data.BaseContainer, *bool)
	var update func()
	var handleevent func(sdl.Event)
	var ok bool

	plug := backend.GetPlugin(path + "/app.so").GetPlugin()

	symInitializer, err := plug.Lookup("Init")
	if err != nil {
		log.Println(err)
	} else {
		initialize, ok = symInitializer.(func(*data.BaseContainer, *bool))
		if !ok {
			log.Fatal("Could not cast Init() function")
		}
	}

	symUpdater, err := plug.Lookup("Update")
	if err != nil {
		log.Println(err)
	} else {
		update, ok = symUpdater.(func())
		if !ok {
			log.Fatal("Could not cast Update() function")
		}
	}

	symHandler, err := plug.Lookup("HandleEvent")
	if err != nil {
		log.Println(err)
	} else {
		handleevent, ok = symHandler.(func(sdl.Event))
		if !ok {
			log.Fatal("Could not cast HandleEvent() function")
		}
	}

	running := true

	handler := normalWindowHandler{window.Parse(), &running, initialize, update, handleevent}

	// initalize here because we didn't have the main container
	// before
	backend.Init(handler.GetContainer())

	// create the window using the handler instance we just declared
	createWindow(handler, &window, windowtype)

	return
}

/*
##############################################################
# Section: Window
##############################################################
*/

type windowHandler interface {
	init(*data.BaseContainer, *bool)
	update()
	handleEvent(sdl.Event)
	GetContainer() *data.BaseContainer
}

func createWindow(handler windowHandler, xmlwindow *parser.XMLWindow, windowtype uint32) (err error) {
	// This variable will will determine wether the window is running or not
	running := true

	// the main container
	// it's a pointer, so any changes should
	// also happen in the handler
	cont := handler.GetContainer()

	position := cont.GetPosition()
	size := cont.GetSize()

	// create an sdl window for the window struct instance
	window, err := sdl.CreateWindow("Sidebar", position.X, position.Y,
		size.X, size.Y, windowtype)
	if err != nil {
		return err
	}

	surface, err := window.GetSurface()
	if err != nil {
		return err
	}
	defer window.Destroy()
	defer sdl.Quit()

	// Initialize the handler
	handler.init(cont, &running)

	// The main loop
	for running {

		// Quit the program in case of exit event
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				fmt.Println("Exit signal received. Quitting...")
				running = false
				break
			default:
				backend.InvokeSDLEvent(event)
				handler.handleEvent(event)
			}
		}

		// parse the container
		// so variables and functions
		// can update
		cont = xmlwindow.Parse()

		handler.update()
		cont.Draw(surface)
		window.UpdateSurface()

	}

	return
}

/*
##############################################################
# Section: Window Handlers
##############################################################
*/

/*
#################################
# "Normal" Window Handler
#################################
*/

type normalWindowHandler struct {
	cont            *data.BaseContainer
	running         *bool
	initializeFunc  func(*data.BaseContainer, *bool)
	updateFunc      func()
	handleeventFunc func(sdl.Event)
}

func (nwh normalWindowHandler) init(c *data.BaseContainer, r *bool) {
	// assign variables to nwh instance
	nwh.cont = c
	nwh.running = r

	nwh.initializeFunc(c, r)
}

func (nwh normalWindowHandler) update() {
	nwh.updateFunc()
}

func (nwh normalWindowHandler) handleEvent(event sdl.Event) {
	nwh.handleeventFunc(event)
}

func (nwh normalWindowHandler) GetContainer() (cont *data.BaseContainer) {
	return nwh.cont
}
