package parser

import (
	"log"
	"plugin"
	"strings"
)

/*
handles app.so data
*/

type Plugin struct {
	plug plugin.Plugin
}

// Open a go plugin file.
// Default file ending is .so
func OpenPluginFile(path string) (plug *Plugin, err error) {
	plugpointer, err := plugin.Open(path)
	// not checking before dereferencing the pointer
	// results in a segmentation violation
	if err != nil {
		return
	}

	return &Plugin{*plugpointer}, err
}

// Gets a variable from the plugin file.
func (p *Plugin) GetVariable(name string) (value string) {
	symVal, err := p.plug.Lookup(name)
	if err != nil {
		log.Fatal(err)
		return
	}

	valuepointer, ok := symVal.(*string)
	if !ok {
		log.Fatal("Could not cast value of variable ", name)
	}

	return *valuepointer
}

// Calls a function inside the plugin file.
// The function has to return a string
func (p *Plugin) CallFunction(name string) (returned string) {
	symFunc, err := p.plug.Lookup(name)
	if err != nil {
		log.Fatal(err)
		return
	}

	value, ok := symFunc.(func() string)
	if !ok {
		log.Fatal("Could not cast value of variable ", name)
	}

	return value()
}

// Gets you a pointer to the plugin
func (p *Plugin) GetPlugin() *plugin.Plugin {
	return &p.plug
}

// Preparses a string
// $ prefix will return the value of a variable with that name
// @ prefix will return the return value of a function with that name
// Everything else will return the original value
func (p *Plugin) PreParseString(str string) (val string) {
	if strings.HasPrefix(str, "$") {
		return p.GetVariable(str[1:])
	} else if strings.HasPrefix(str, "@") {
		return p.CallFunction(str[1:])
	}

	return str
}
