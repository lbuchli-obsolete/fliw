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
func (p *Plugin) CallFunction(function string) (returned string) {
	var name string
	var args []string

	// separate function name and function arguments
	for i, c := range function {
		if c == '(' {
			name = function[:i]
			args = strings.Split(function[i+1:len(function)-1], ",")
			break
		}
	}

	// clean up arguments (in case there are spaces before the name)
	// and evaluate them
	for i, arg := range args {
		for strings.HasPrefix(arg, " ") {
			arg = arg[1:]
		}

		// if the argument is empty, remove it from the list
		// and skip it
		if arg == "" {
			args = append(args[:i], args[i+1:]...)
			continue
		}

		// if there are any operands in the string
		// or the string starts with brackets
		if strings.ContainsAny(arg, "+-*/^") || string(arg[0]) == "(" {
			// evaluate the argument
			args[i] = CalculateValue(arg)
		}
	}

	symFunc, err := p.plug.Lookup(name)
	if err != nil {
		log.Fatal(err)
		return
	}

	if len(args) > 0 {
		value, ok := symFunc.(func(...string) string)
		if !ok {
			value, ok := symFunc.(func(...string))
			if !ok {
				log.Fatal("Could not cast function with arguments: ", name)
				return ""
			}

			value(args...)
			return ""
		}
		return value(args...)

	} else {
		value, ok := symFunc.(func() string)
		if !ok {
			value, ok := symFunc.(func())
			if !ok {
				log.Fatal("Could not cast function: ", name)
				return ""
			}

			value()
			return ""
		}

		return value()
	}
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
	// return if string is empty
	if str == "" {
		return ""
	}

	if string(str[0]) == "$" {
		return p.GetVariable(str[1:])
	} else if string(str[0]) == "@" {
		return p.CallFunction(str[1:])
	}

	return str
}
