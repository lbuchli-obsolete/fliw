# FLIW

this project is not in a usable state yet.

## Framework for Low Interaction Windows

This project originated from my swayland project. It is a framework for creating
Window appearance & Window behavior with next to no user interaction (e.g. keyboard, mouse, ...)

FLIW is written in go. All describing of window behavior is also done in go, while the look of
the window is described in XML. Thus, a basic module structure may look like this:


MODULE/

	style.xml <- describes the look of the window

	app.go <- source for the behavior part

	app.so <- behavior part compiled from module.go as plugin (go build -buildmode=plugin)

	README.md <- description of your module



## Installation

TODO write installation guide
