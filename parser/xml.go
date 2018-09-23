package parser

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/phoenixdevelops/fliw/backend"
	"github.com/phoenixdevelops/fliw/data"
	"github.com/phoenixdevelops/fliw/image"
	xsdvalidate "github.com/terminalstatic/go-xsd-validate"
	"github.com/veandco/go-sdl2/sdl"
)

/*
parses your style.xml data
*/

// WindowType lists all the window types
var WindowType = map[string]uint32{
	"always_on_top":      sdl.WINDOW_ALWAYS_ON_TOP,
	"allow_highdpi":      sdl.WINDOW_ALLOW_HIGHDPI,
	"borderless":         sdl.WINDOW_BORDERLESS,
	"foreign":            sdl.WINDOW_FOREIGN,
	"fullscreen":         sdl.WINDOW_FULLSCREEN,
	"fullscreen_desktop": sdl.WINDOW_FULLSCREEN_DESKTOP,
	"hidden":             sdl.WINDOW_HIDDEN,
	"input_focus":        sdl.WINDOW_INPUT_FOCUS,
	"input_grabbed":      sdl.WINDOW_INPUT_GRABBED,
	"maximized":          sdl.WINDOW_MAXIMIZED,
	"minimized":          sdl.WINDOW_MINIMIZED,
	"mouse_capture":      sdl.WINDOW_MOUSE_CAPTURE,
	"mouse_focus":        sdl.WINDOW_MOUSE_FOCUS,
	"opengl":             sdl.WINDOW_OPENGL,
	"popup_menu":         sdl.WINDOW_POPUP_MENU,
	"resizable":          sdl.WINDOW_RESIZABLE,
	"shown":              sdl.WINDOW_SHOWN,
	"skip_taskbar":       sdl.WINDOW_SKIP_TASKBAR,
	"tooltip":            sdl.WINDOW_TOOLTIP,
	"utility":            sdl.WINDOW_UTILITY,
	"vulkan":             sdl.WINDOW_VULKAN,
}

// XMLItem is an interface for an XML item.
// It must be parsable to its data.Item equivalent
type XMLItem interface {
	parse(data.Vector, string) data.Item
	isStatic(string) bool
	getUID() uint
}

// XMLContainer is an extension to the XML item interface
// allowing a parseToCont function
type XMLContainer interface {
	XMLItem
	parseToCont(data.Vector, string) data.Container
}

// XMLItem is the base of all XML elements
// it defines basic things like position, size and events
type XMLBase struct {
	UID     uint
	X       string `xml:"x,attr"`
	Y       string `xml:"y,attr"`
	Width   string `xml:"width,attr"`
	Height  string `xml:"height,attr"`
	OnEvent string `xml:"onevent,attr"`
	Static  string `xml:"static,attr"`
}

type XMLContainerBase struct {
	XMLBase
	Color     string             `xml:"color,attr"`
	Conts     []XMLBaseContainer `xml:"container"`
	ListConts []XMLListContainer `xml:"listcontainer"`
	Labels    []XMLLabel         `xml:"label"`
	Textures  []XMLTexture       `xml:"texture"`
	Unicolors []XMLUnicolor      `xml:"unicolor"`
	Links     []XMLLink          `xml:"link"`
}

func (base XMLBase) isStatic(plugin string) bool {
	return parseBool(base.Static, plugin)
}

func (base XMLBase) getUID() uint {
	return base.UID
}

// XMLWindow is the base XML element of each
// style.xml file
type XMLWindow struct {
	XMLName    xml.Name `xml:"window"`
	WindowType string   `xml:"windowtype,attr"`
	XMLBaseContainer
}

// XMLExtension is the base element of each
//linked XML file
type XMLExtension struct {
	XMLName xml.Name `xml:"extension"`
	Backend string   `xml:"backend,attr"`
	XMLBaseContainer
}

// XMLBaseContainer is an item that holds other items.
type XMLBaseContainer struct {
	XMLName xml.Name `xml:"container"`
	XMLContainerBase
}

// XMLListContainer is a XMLBaseContainer but it stacks
// the items ontop of each other when being drawn
type XMLListContainer struct {
	XMLName xml.Name `xml:"listcontainer"`
	XMLContainerBase
}

// XMLLabel is an item displaying basic short text
type XMLLabel struct {
	XMLName xml.Name `xml:"label"`
	XMLBase
	TextSize string `xml:"textsize,attr"`
	VAlign   string `xml:"valign,attr"`
	HAlign   string `xml:"halign,attr"`
	FGColor  string `xml:"fgcolor,attr"`
	BGColor  string `xml:"bgcolor,attr"`
	Bold     string `xml:"bold,attr"`
	Text     string `xml:",chardata"`
}

// XMLTexture is an item displaying a texture (picture)
type XMLTexture struct {
	XMLName xml.Name `xml:"texture"`
	XMLBase
	Texture   string `xml:",chardata"`
	ScaleDown string `xml:"scaledown,attr"`
}

// XMLUnicolor is an item displaying only a single color
type XMLUnicolor struct {
	XMLName xml.Name `xml:"unicolor"`
	XMLBase
	Color string `xml:",chardata"`
}

// XMLLink links to another XML file.
// It has no data counterpart
type XMLLink struct {
	XMLName xml.Name `xml:"link"`
	XMLBase
	Link string `xml:",chardata"`
}

/*
###########################################
# Parser and window information
###########################################
*/

var bgcolor uint32
var bounds sdl.Rect
var dirpath string

var staticItems map[uint]*data.Item

var uidIndex uint

// UnmarshalXMLFile gives back a XMLWIndow object which can be parsed further
func UnmarshalXMLFile(path string) (window XMLWindow, windowtype uint32, err error) {
	dirpath = path
	staticItems = make(map[uint]*data.Item)

	// open the file
	file, err := ioutil.ReadFile(path + "/style.xml")
	if err != nil {
		return
	}

	// validate xml file.
	// it will be validated against
	// assets/style.xsd
	ok, err := validateXMLFile(file, "assets/style.xsd")
	if !ok {
		log.Println("XML config contained errors.")
	}
	if err != nil {
		log.Fatal(err)
	}

	win := XMLWindow{
		WindowType: "popup_menu",
	}

	// unmarshal the file
	err = xml.Unmarshal([]byte(file), &win)
	if err != nil {
		return
	}

	win.XMLBaseContainer.assignUIDs()

	// get the display size
	bounds, err = sdl.GetDisplayBounds(0)
	if err != nil {
		return
	}

	// default to sdl.WINDOW_SHOWN
	if win.WindowType == "" {
		return win, WindowType["shown"], err
	}

	return win, WindowType[win.WindowType], err
}

func validateXMLFile(file []byte, xsdpath string) (valid bool, err error) {
	// validate xml with an xsd file
	xsdvalidate.Init()
	defer xsdvalidate.Cleanup()
	xsdhandler, err := xsdvalidate.NewXsdHandlerUrl(xsdpath, xsdvalidate.ParsErrDefault)
	if err != nil {
		return true, err
	}
	defer xsdhandler.Free()

	xmlhandler, err := xsdvalidate.NewXmlHandlerMem(file, xsdvalidate.ParsErrDefault)
	if err != nil {
		return true, err
	}
	defer xmlhandler.Free()

	err = xsdhandler.Validate(xmlhandler, xsdvalidate.ValidErrDefault)
	if err != nil {
		switch err.(type) {
		case xsdvalidate.ValidationError:
			// there was an error in the file, not our fault
			log.Println(err)
			log.Printf("Error in line: %d\n", err.(xsdvalidate.ValidationError).Errors[0].Line)
			log.Println(err.(xsdvalidate.ValidationError).Errors[0].Message)
			return false, nil
		default:
			// our fault?
			return true, err
		}
	}

	return true, nil
}

// parse any item to its data.Item counterpart
func parseItem(item XMLItem, psize data.Vector, plugin string) data.Item {
	if item.isStatic(plugin) {
		if val, ok := staticItems[item.getUID()]; ok {
			return *val
		} else {
			val := item.parse(psize, plugin)
			staticItems[item.getUID()] = &val
			return val
		}
	}

	return item.parse(psize, plugin)
}

// recursively assigns a UID to the container and all its children
func (cont *XMLContainerBase) assignUIDs() {
	// assign own uid
	cont.UID = uidIndex
	uidIndex++

	// assign all the normal items a uid
	for i := range cont.Labels {
		cont.Labels[i].UID = uidIndex
		uidIndex++
	}
	for i := range cont.Textures {
		cont.Textures[i].UID = uidIndex
		uidIndex++
	}
	for i := range cont.Unicolors {
		cont.Unicolors[i].UID = uidIndex
		uidIndex++
	}
	for i := range cont.Links {
		cont.Links[i].UID = uidIndex
		uidIndex++
	}

	// let the containers assign uids
	for i := range cont.Conts {
		cont.Conts[i].assignUIDs()
		uidIndex++
	}
	for i := range cont.ListConts {
		cont.ListConts[i].assignUIDs()
		uidIndex++
	}
}

/*
###########################################
# Struct Parser
###########################################
*/

// Parse gets a drawable data.Container from an XMLWindow
func (win *XMLWindow) Parse() (maincont data.Container) {
	bgcolor = parseColor(win.Color, dirpath+"/app.so")
	return win.parseToCont(data.Vector{X: bounds.W, Y: bounds.H}, dirpath+"/app.so")
}

// converts XMLContainer to data.Container
func (cont XMLBaseContainer) parseToCont(psize data.Vector, plugin string) (container data.Container) {
	// get the size
	size := parseWH(cont.Width, cont.Height, psize, plugin)

	// the map of items any data.Container contains
	itemmap := make(map[string]data.Item)

	// Add the items to the list
	for it, item := range cont.Labels {
		itemmap["label"+strconv.Itoa(it)] = parseItem(item, size, plugin)
	}
	for it, item := range cont.Textures {
		itemmap["texture"+strconv.Itoa(it)] = parseItem(item, size, plugin)
	}
	for it, item := range cont.Unicolors {
		itemmap["unicolor"+strconv.Itoa(it)] = parseItem(item, size, plugin)
	}
	for it, item := range cont.Conts {
		itemmap["container"+strconv.Itoa(it)] = parseItem(item, size, plugin)
	}
	for it, item := range cont.ListConts {
		itemmap["listcontainer"+strconv.Itoa(it)] = parseItem(item, size, plugin)
	}
	for it, item := range cont.Links {
		itemmap["link"+strconv.Itoa(it)] = parseItem(item, size, plugin)
	}

	// Construct container
	return &data.BaseContainer{
		Position: parseXY(cont.X, cont.Y, psize, plugin),
		Size:     size,
		BGcolor:  parseColor(cont.Color, plugin),
		Items:    itemmap,
		Events:   parseEvents(cont.OnEvent, plugin),
		IsLink:   false,
	}
}

// converts XMLContainer to data.Item
func (cont XMLBaseContainer) parse(psize data.Vector, plugin string) (container data.Item) {
	return cont.parseToCont(psize, plugin)
}

// converts XMLListContainer to data.Container
func (cont XMLListContainer) parseToCont(psize data.Vector, plugin string) (listcontainer data.Container) {
	// get the size
	size := parseWH(cont.Width, cont.Height, psize, plugin)

	// the map of items any data.Container contains
	itemmap := make(map[string]data.ItemEntry)
	itemindex := 0

	// Add the items to the list
	for it, item := range cont.Labels {
		itemmap["label"+strconv.Itoa(it)] = data.ItemEntry{Index: itemindex, Item: parseItem(item, size, plugin)}
		itemindex++
	}
	for it, item := range cont.Textures {
		itemmap["texture"+strconv.Itoa(it)] = data.ItemEntry{Index: itemindex, Item: parseItem(item, size, plugin)}
		itemindex++
	}
	for it, item := range cont.Unicolors {
		itemmap["unicolor"+strconv.Itoa(it)] = data.ItemEntry{Index: itemindex, Item: parseItem(item, size, plugin)}
		itemindex++
	}
	for it, item := range cont.Conts {
		itemmap["container"+strconv.Itoa(it)] = data.ItemEntry{Index: itemindex, Item: parseItem(item, size, plugin)}
		itemindex++
	}
	for it, item := range cont.ListConts {
		itemmap["listcontainer"+strconv.Itoa(it)] = data.ItemEntry{Index: itemindex, Item: parseItem(item, size, plugin)}
		itemindex++
	}
	for it, item := range cont.Links {
		itemmap["link"+strconv.Itoa(it)] = data.ItemEntry{Index: itemindex, Item: parseItem(item, size, plugin)}
		itemindex++
	}

	// Construct container
	return &data.ListContainer{
		Position: parseXY(cont.X, cont.Y, psize, plugin),
		Size:     size,
		BGcolor:  parseColor(cont.Color, plugin),
		Items:    itemmap,
		Events:   parseEvents(cont.OnEvent, plugin),
		IsLink:   false,
	}
}

// converts XMLListContainer to data.Item
func (cont XMLListContainer) parse(psize data.Vector, plugin string) (listcontainer data.Item) {
	return cont.parseToCont(psize, plugin)
}

// converts XMLUnicolor to data.Unicolor
func (uni XMLUnicolor) parse(psize data.Vector, plugin string) (unicolor data.Item) {
	// Construct Unicolor
	return &data.Unicolor{
		Position: parseXY(uni.X, uni.Y, psize, plugin),
		Size:     parseWH(uni.Width, uni.Height, psize, plugin),
		Color:    parseColor(uni.Color, plugin),
		Events:   parseEvents(uni.OnEvent, plugin),
	}
}

// converts XMLLabel to data.Label
func (lab XMLLabel) parse(psize data.Vector, plugin string) (label data.Item) {
	// Construct Label
	return &data.Label{
		Position: parseXY(lab.X, lab.Y, psize, plugin),
		Size:     parseWH(lab.Width, lab.Height, psize, plugin),
		Text:     parseText(lab.Text, plugin),
		Textsize: parseInt(lab.TextSize, plugin),
		Valign:   parseAlign(lab.VAlign, plugin),
		Halign:   parseAlign(lab.HAlign, plugin),
		Color:    parseColor(lab.FGColor, plugin),
		BGcolor:  parseColor(lab.BGColor, plugin),
		Bold:     parseBool(lab.Bold, plugin),
		Events:   parseEvents(lab.OnEvent, plugin),
	}
}

// converts XMLTexture to data.Texture
func (tex XMLTexture) parse(psize data.Vector, plugin string) (texture data.Item) {
	// Construct Texture
	return &data.Texture{
		Position: parseXY(tex.X, tex.Y, psize, plugin),
		Size:     parseWH(tex.Width, tex.Height, psize, plugin),
		Texture: parseImage(tex.Texture, parseXY(tex.Width, tex.Height, psize, plugin),
			parseBool(tex.ScaleDown, plugin), plugin),
		Events: parseEvents(tex.OnEvent, plugin),
	}
}

var links = make(map[string]*XMLExtension)

// converts XMLLink to data.Container
func (link XMLLink) parse(psize data.Vector, plugin string) (cont data.Item) {
	filepath := parsePath(link.Link, plugin)

	ext, ok := links[filepath]
	var newplug string
	// if extension wasn't read already
	if !ok {
		// read the extension file
		file, err := ioutil.ReadFile(filepath)
		if err != nil {
			log.Fatal(err)
			return
		}

		// validate the extension file
		valid, err := validateXMLFile(file, "assets/extension.xsd")
		if err != nil {
			log.Fatal(err)
			return
		}
		if !valid {
			log.Fatal("XML extension file ", filepath, " is not valid.")
			return
		}

		// unmarshal the extension file
		ext = &XMLExtension{}
		err = xml.Unmarshal(file, ext)
		if err != nil {
			log.Fatal(err)
			return
		}

		// assign uids
		ext.XMLBaseContainer.assignUIDs()

		// save work for later
		links[filepath] = ext

		newplug = parsePath(ext.Backend, plugin)

		// register backend
		backend.AddPlugin(newplug)
	} else {
		newplug = parsePath(ext.Backend, plugin)
	}

	// needs to be set so child elements can
	// base their size on it
	ext.XMLBaseContainer.Width = link.Width
	ext.XMLBaseContainer.Height = link.Height

	datacont := ext.XMLBaseContainer.parseToCont(psize, newplug)

	// overwrite x, y, width, height
	datacont.SetPosition(parseXY(link.X, link.Y, psize, plugin))
	datacont.SetSize(parseWH(link.Width, link.Height, psize, plugin))

	// make the container a link
	datacont.SetLink(true)

	return datacont
}

/*
###########################################
# String Parser
###########################################
*/

// parses a string containing a hex color to a uint32
// representing said color
// defaults to bgcolor
// if you want sdl to draw the right color, you'll have to use parseColor(),
// which does the same exept it swaps some bytes
func parseColor(color string, plugin string) (result uint32) {
	// return default if not specified
	if color == "" {
		return bgcolor
	}

	// remove whitespaces
	color = cleanString(color)

	// preprocess
	color = backend.GetPlugin(plugin).PreParseString(color)

	if strings.HasPrefix(color, "#") {
		// Remove # prefix
		color = color[1:]
	}

	// decode hex string to byte array
	val, err := hex.DecodeString(color)
	if err != nil {
		log.Fatal(err)
		return bgcolor
	}

	// if no alpha value specified
	if len(val) == 3 {
		val = append(val, 0)
	}

	// return the byte array as uint32 (use imgtools to swap color into correct order)
	return binary.LittleEndian.Uint32(val)
}

// parses a string to a bool value. Defaults to false if string is empty
func parseBool(b string, plugin string) (result bool) {
	b = cleanString(b)

	// preprocess
	b = backend.GetPlugin(plugin).PreParseString(b)

	switch b {
	case "false":
		return false
	case "true":
		return true
	case "": // default false for empty string
		return false
	default:
		log.Fatal(errors.New("Invalid boolean value: " + b))
		return false
	}
}

// parses a string to a data.Align value. Defaults to CENTER if string is empty
func parseAlign(align string, plugin string) (result data.Align) {
	align = cleanString(align)

	// preprocess
	align = backend.GetPlugin(plugin).PreParseString(align)

	switch align {
	case "top":
		return data.TOP
	case "center":
		return data.CENTER
	case "bottom":
		return data.BOTTOM
	case "right":
		return data.RIGHT
	case "left":
		return data.LEFT
	case "": // default center for empty string
		return data.CENTER
	default:
		log.Fatal(errors.New("Invalid Align Value: " + align))
		return data.CENTER
	}
}

// parses x and y strings to a data.Vector position. Defaults to 0 if string is empty.
// for width and height use parseWH.
// Uses parentSize for percentual interpretation
func parseXY(x string, y string, parentSize data.Vector, plugin string) (result data.Vector) {
	x = cleanString(x)
	y = cleanString(y)

	plug := backend.GetPlugin(plugin)

	// preprocess
	x = plug.PreParseString(x)
	y = plug.PreParseString(y)
	x = preparseNumberString(x, plugin)
	y = preparseNumberString(y, plugin)

	// function for parsing 1 dimension
	strparse := func(v string, pv int32) (res int32) {
		if strings.HasSuffix(v, "%") {
			// get the int value of the string without the trailing '%'
			vi, e := strconv.Atoi(v[:len(v)-1])
			if e != nil {
				log.Fatal(e)
				return res
			}

			return int32((float32(vi) / float32(100)) * float32(pv))
		} else if v == "" {
			return 0
		} else {
			// normal, no percentage
			vi, e := strconv.Atoi(v)
			if e != nil {
				log.Fatal(e)
				return res
			}

			return int32(vi)
		}
	}

	// calculate x
	xi := strparse(x, parentSize.X)

	// calculate y
	yi := strparse(y, parentSize.Y)

	return data.Vector{X: xi, Y: yi}
}

// same as parseXY, except it defaults to 100%
func parseWH(w string, h string, parentSize data.Vector, plugin string) (result data.Vector) {
	result = parseXY(w, h, parentSize, plugin)

	if result.X == 0 {
		result.X = parentSize.X
	}
	if result.Y == 0 {
		result.Y = parentSize.Y
	}

	return
}

// parses a string to an int. Defaults to 0 if empty
func parseInt(integer string, plugin string) (result int) {
	integer = cleanString(integer)

	// preprocess
	integer = backend.GetPlugin(plugin).PreParseString(integer)

	integer = preparseNumberString(integer, plugin)

	if integer == "" {
		return 0
	}

	result, err := strconv.Atoi(integer)
	if err != nil {
		log.Fatal(err)
		return result
	}

	return result
}

// a map of images already loaded and converted to a surface
var loadedimages = make(map[string]*sdl.Surface)

// parses a string (path) to an *sdl.Surface. Defaults to a unicolored surface if empty
func parseImage(imagepath string, size data.Vector, downscale bool, plugin string) (result *sdl.Surface) {
	imagepath = parsePath(imagepath, plugin)

	// return the already processed value of that image
	// if existent
	if val, ok := loadedimages[imagepath]; ok {
		return val
	}

	// get image object from file
	img, err := image.FromFile(imagepath)
	if err != nil {
		log.Fatal(err)
		result, err := sdl.CreateRGBSurface(0, size.X, size.Y, 32, 0, 0, 0, 0)
		if err != nil {
			log.Fatal(err)
			return result
		}

		result.FillRect(nil, bgcolor)
		return result
	}

	// convert image object to surface
	surf, err := image.ImgToSurface(img)
	if err != nil {
		log.Fatal(err)
		result, err := sdl.CreateRGBSurface(0, size.X, size.Y, 32, 0, 0, 0, 0)
		if err != nil {
			log.Fatal(err)
			return result
		}

		result.FillRect(nil, bgcolor)
		return result
	}

	// scale surface if needed
	if downscale {
		result, err = image.ResizeSurface(surf, size.X, size.Y)
		if err != nil {
			log.Fatal(err)
			result, err := sdl.CreateRGBSurface(0, size.X, size.Y, 32, 0, 0, 0, 0)
			if err != nil {
				log.Fatal(err)
				return result
			}

			result.FillRect(nil, bgcolor)
			return result
		}
	} else {
		result = surf
	}

	// also save the image in a map of already loaded images
	loadedimages[imagepath] = result

	return result
}

// parses a string to a string (removes whitespace before and after)
func parseText(text string, plugin string) (result string) {
	text = cleanString(text)
	return backend.GetPlugin(plugin).PreParseString(text)
}

// parses an ItemEvents struct to a map of events understood by the data package
// Entries may look like this:
// onevent="click:func1,rightclick:func2"
func parseEvents(onevent string, plugin string) (result map[string]string) {
	result = make(map[string]string)

	onevent = cleanString(onevent)

	entries := strings.Split(onevent, ",")

	for _, entry := range entries {
		data := strings.Split(entry, ":")

		if len(data) == 2 {
			data[0] = parseText(data[0], plugin)
			data[1] = parseText(data[1], plugin)
			result[data[0]] = plugin + "/" + data[1]
		}
	}

	return
}

// parses  a string containing a path
// into an absolute path
func parsePath(path string, plugin string) (result string) {
	path = cleanString(path)

	// preprocess
	path = backend.GetPlugin(plugin).PreParseString(path)

	if !strings.HasPrefix(path, "/") {
		return dirpath + "/" + path
	}

	return path
}

// cleans up whitespace before and after a string
func cleanString(s string) (cleaned string) {
	if s == "" {
		return s
	}

	// remove whitespaces before
	for strings.ContainsAny(string(s[0]), "\a\f\t\n\r\v") {
		s = s[1:]
	}

	// remove whitespaces after
	for strings.ContainsAny(string(s[len(s)-1]), "\a\f\t\n\r\v") {
		s = s[:len(s)-1]
	}

	return s
}

// preparses a string containing numbers or operations
// if there is something to be calculated, e.g. 3*2.5
// it will be done here
func preparseNumberString(s string, plugin string) (result string) {
	// if the string is empty, give back an empty string
	if s == "" {
		return ""
	}

	// if there are any operands in the string
	// or the string starts with brackets
	if strings.ContainsAny(s, "+-*/^") || string(s[0]) == "(" {
		return backend.CalculateValue(s, backend.GetPlugin(plugin))
	}

	return s
}
