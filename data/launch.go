package data

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

/*
launches your app using your apps root directory.
*/

var display_size Vector

/*
###############################################################
# Section: Initialization
###############################################################
*/

func initialize() {
	// Initialize sdl.sdl and sdl.ttf
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ttf.Init()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Get the display size
	bounds, err := sdl.GetDisplayBounds(0)
	if err != nil {
		fmt.Println(err)
		return
	}

	display_size = Vector{bounds.W, bounds.H}
}

/*
##############################################################
# Section: Window
##############################################################
*/

type WindowHandler interface {
	init(*Container, *bool)
	update()
	handleEvent(sdl.Event)
}

func createWindow(position Vector, size Vector, bgcolor uint32, handler WindowHandler) (err error) {
	// This variable will will determine wether the window is running or not
	running := true

	// the main container
	cont := Container{Vector{0, 0}, size, bgcolor, make(map[string]Item)}

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
	handler.init(&cont, &running)

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
	cont    *Container
	running *bool
}

func (nwh NormalWindowHandler) init(c *Container, r *bool) {
	// assign variables to nwh instance
	nwh.cont = c
	nwh.running = r

}

func (nwh NormalWindowHandler) update() {
	return
}

func (nwh NormalWindowHandler) handleEvent(event sdl.Event) {
	return
}

/*
##############################################################
# Section: Parser
##############################################################
*/
