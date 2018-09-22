package backend

import (
	"log"
	"regexp"
	"runtime/debug"

	"github.com/phoenixdevelops/fliw/data"
	"github.com/phoenixdevelops/fliw/input"
	"github.com/veandco/go-sdl2/sdl"
)

// EventName is a string containing the name of an event
type EventName string

// event names
const (
	MouseclickEvent      EventName = "mouseclick"
	MouserightclickEvent EventName = "mouserightclick"
	MousereleaseEvent    EventName = "mouserelease"
	KeydownEvent         EventName = "keydown"
	KeyupEvent           EventName = "keyup"
)

// Event struct describes an event
type Event struct {
	Name          EventName
	MousePosition data.Vector
}

var plugins map[string]*Plugin = make(map[string]*Plugin)
var baseContainer *data.BaseContainer

var pathRegex *regexp.Regexp

// Init should be called before using any other funtions
// in this package
func Init(basecontainer *data.BaseContainer) {
	baseContainer = basecontainer

	pathRegex = regexp.MustCompile(`([[:graph:]]+/)+`)
}

// Invoke invokes an event
// and calls the function in the right plugin files
func Invoke(event Event) {
	// make a deep copy of the container
	var bcont data.BaseContainer = *baseContainer
	var container data.Container = &bcont

	// the main container gets all events
	if ev := container.GetEvent(string(event.Name)); ev != "" {
		callFunction(ev)
	}

	for true {
		item := container.GetItemAt(event.MousePosition)

		// if the item is of type container
		if val, ok := item.(data.Container); ok {
			// if they are the same
			if val == container {
				if ev := item.GetEvent(string(event.Name)); ev != "" {
					callFunction(ev)
				}
				return
			}

			// if the container is a link also call the event in the container
			if container.GetIsLink() {
				if ev := container.GetEvent(string(event.Name)); ev != "" {
					callFunction(ev)
				}
			}

			container = val
			event.MousePosition.X -= item.GetPosition().X
			event.MousePosition.Y -= item.GetPosition().Y

		} else {
			if ev := item.GetEvent(string(event.Name)); ev != "" {
				callFunction(ev)
			}
			return
		}
	}
}

// AddPlugin adds a plugin file
// if the specified path was not added before
func AddPlugin(path string) {
	if _, ok := plugins[path]; !ok {
		plugin, err := OpenPluginFile(path)
		if err != nil {
			log.Fatal(err)
		}

		plugins[path] = plugin
	}
}

// GetPlugin gets you the plugin matching a go .so file
func GetPlugin(path string) (plugin *Plugin) {
	if path == "" {
		debug.PrintStack()
		log.Fatal("Tried to get plugin from empty string")
	}

	if result, ok := plugins[path]; ok {
		return result
	}

	debug.PrintStack()
	log.Fatal("Tried to get plugin where no plugin was registered: ", path)
	return
}

var isMouseClicked = false

// InvokeSDLEvent invokes an sdl event
func InvokeSDLEvent(event sdl.Event) {
	switch t := event.(type) {
	case *sdl.MouseButtonEvent:
		if t.Button == sdl.BUTTON_LEFT && !isMouseClicked {
			isMouseClicked = true
			Invoke(Event{MouseclickEvent, data.Vector{X: t.X, Y: t.Y}})
		} else if t.Button == sdl.BUTTON_RIGHT && !isMouseClicked {
			isMouseClicked = true
			Invoke(Event{MouserightclickEvent, data.Vector{X: t.X, Y: t.Y}})
		} else if isMouseClicked {
			isMouseClicked = false
			Invoke(Event{MousereleaseEvent, data.Vector{X: t.X, Y: t.Y}})
		}
	case *sdl.KeyboardEvent:
		if t.GetType() == sdl.KEYDOWN {
			if isevent := input.PressKey(t.Keysym.Sym); isevent {
				x, y, _ := sdl.GetMouseState()
				Invoke(Event{KeydownEvent, data.Vector{X: x, Y: y}})
			}
		} else if t.GetType() == sdl.KEYUP {
			if isevent := input.PressKey(t.Keysym.Sym); isevent {
				x, y, _ := sdl.GetMouseState()
				Invoke(Event{KeyupEvent, data.Vector{X: x, Y: y}})
			}
		}
	}
}

// calls a function based on its path and its name
func callFunction(function string) {
	path := pathRegex.FindString(function)

	plugins[path[:len(path)-1]].CallFunction(function[len(path):])
}
