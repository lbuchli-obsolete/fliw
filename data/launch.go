package data

import (
	"fmt"
	"log"
	"plugin"

	"github.com/veandco/go-sdl2/sdl"
)

/*
launches your app using your apps root directory.
*/

/*
##############################################################
# Section: Window
##############################################################
*/

type WindowHandler interface {
	init(*BaseContainer, *bool)
	update()
	handleEvent(sdl.Event)
	GetContainer() *BaseContainer
}

func createWindow(position Vector, size Vector, bgcolor uint32, handler WindowHandler) (err error) {
	// This variable will will determine wether the window is running or not
	running := true

	// the main container
	cont := handler.GetContainer()

	// create an sdl window for the window struct instance
	window, err := sdl.CreateWindow("Sidebar", position.X, position.Y,
		cont.Size.X, cont.Size.Y, sdl.WINDOW_POPUP_MENU)
	if err != nil {
		return err
	}

	surface, err := window.GetSurface()
	if err != nil {
		return err
	}
	defer window.Destroy()
	defer sdl.Quit()

	// Set the background color
	surface.FillRect(nil, bgcolor)

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
				handler.handleEvent(event)
			}
		}

		handler.update()
		cont.Draw(surface)
		window.UpdateSurface()
	}

	return nil
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

type NormalWindowHandler struct {
	cont        *BaseContainer
	running     *bool
	initializer Initializer
	updater     Updater
	handler     EventHandler
}

func (nwh NormalWindowHandler) init(c *BaseContainer, r *bool) {
	// assign variables to nwh instance
	nwh.cont = c
	nwh.running = r

	nwh.initializer.Initialize(c, r)
}

func (nwh NormalWindowHandler) update() {
	nwh.updater.Update()
}

func (nwh NormalWindowHandler) handleEvent(event sdl.Event) {
	nwh.handler.HandleEvent(event)
}

func (nwh NormalWindowHandler) GetContainer() (cont *BaseContainer) {
	return nwh.cont
}

/*
##############################################################
# Section: Launcher
##############################################################
*/

type Initializer interface {
	Initialize(*BaseContainer, *bool)
}

type Updater interface {
	Update()
}

type EventHandler interface {
	HandleEvent(sdl.Event)
}

// Empty types as fallback default

type DefaultInitializer struct{}

func (di DefaultInitializer) Initialize(cont *BaseContainer, running *bool) {
	return
}

type DefaultUpdater struct{}

func (du DefaultUpdater) Update() {
	return
}

type DefaultEventHandler struct{}

func (deh DefaultEventHandler) HandleEvent(event sdl.Event) {
	return
}

func ShowWindow(basecont *BaseContainer, bgcolor uint32, plug plugin.Plugin) {
	// Get a function for initializing, updating and eventhandling

	var initializer Initializer
	var updater Updater
	var eventhandler EventHandler
	var ok bool

	// get the initializer type
	symInitializer, err := plug.Lookup("Initializer")
	if err != nil {
		log.Println(err)
		initializer = DefaultInitializer{}
	} else {
		initializer, ok = symInitializer.(Initializer)
		if !ok {
			initializer = DefaultInitializer{}
		}
	}

	// get the updater type
	symUpdater, err := plug.Lookup("Updater")
	if err != nil {
		log.Println(err)
		updater = DefaultUpdater{}
	} else {
		updater, ok = symUpdater.(Updater)
		if !ok {
			updater = DefaultUpdater{}
		}
	}

	// get the eventhandler type
	symHandler, err := plug.Lookup("EventHandler")
	if err != nil {
		log.Println(err)
		eventhandler = DefaultEventHandler{}
	} else {
		eventhandler, ok = symHandler.(EventHandler)
		if !ok {
			eventhandler = DefaultEventHandler{}
		}
	}

	running := true

	handler := NormalWindowHandler{basecont, &running, initializer, updater, eventhandler}

	// create the window using the handler instance we just declared
	createWindow(basecont.GetPosition(), basecont.GetSize(), bgcolor, handler)
}
