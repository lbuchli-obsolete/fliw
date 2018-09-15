package launch

import (
	"fmt"
	"log"

	"github.com/phoenixdevelops/fliw/data"
	"github.com/phoenixdevelops/fliw/parser"
	"github.com/veandco/go-sdl2/sdl"
)

/*
launches your app using your apps root directory.
*/

/*
##############################################################
# Section: Launcher
##############################################################
*/

type Initializer interface {
	Initialize(*data.BaseContainer, *bool)
}

type Updater interface {
	Update()
}

type EventHandler interface {
	HandleEvent(sdl.Event)
}

// Empty types as fallback default

type DefaultInitializer struct{}

func (di DefaultInitializer) Initialize(cont *data.BaseContainer, running *bool) {
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

// Shows a window given the path to the config files
func ShowWindow(path string) (err error) {
	// process the files
	// these steps are fatal:
	// if an error occurs, the program can't continue

	plugin, err := parser.OpenPluginFile(path + "/app.so")
	if err != nil {
		return
	}

	window, windowtype, err := parser.UnmarshalXMLFile(path+"/style.xml", plugin)
	if err != nil {
		return
	}

	// Get a function for initializing, updating and eventhandling

	var initializer Initializer
	var updater Updater
	var eventhandler EventHandler
	var ok bool

	plug := plugin.GetPlugin()

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

	handler := NormalWindowHandler{window.Parse(), &running, initializer, updater, eventhandler}

	// create the window using the handler instance we just declared
	createWindow(handler, &window, windowtype)

	return
}

/*
##############################################################
# Section: Window
##############################################################
*/

type WindowHandler interface {
	init(*data.BaseContainer, *bool)
	update()
	handleEvent(sdl.Event)
	GetContainer() *data.BaseContainer
}

func createWindow(handler WindowHandler, xmlwindow *parser.XMLWindow, windowtype uint32) (err error) {
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

type NormalWindowHandler struct {
	cont        *data.BaseContainer
	running     *bool
	initializer Initializer
	updater     Updater
	handler     EventHandler
}

func (nwh NormalWindowHandler) init(c *data.BaseContainer, r *bool) {
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

func (nwh NormalWindowHandler) GetContainer() (cont *data.BaseContainer) {
	return nwh.cont
}
