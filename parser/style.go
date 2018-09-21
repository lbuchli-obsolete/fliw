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

	"github.com/phoenixdevelops/fliw/data"
	"github.com/phoenixdevelops/fliw/imgtools"
	xsdvalidate "github.com/terminalstatic/go-xsd-validate"
	"github.com/veandco/go-sdl2/sdl"
)

/*
parses your style.xml data
*/

// all the window types
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

type XMLItem struct {
	X       string `xml:"x,attr"`
	Y       string `xml:"y,attr"`
	Width   string `xml:"width,attr"`
	Height  string `xml:"height,attr"`
	OnEvent string `xml:"onevent,attr"`
}

type XMLWindow struct {
	XMLName    xml.Name `xml:"window"`
	WindowType string   `xml:"windowtype,attr"`
	XMLBaseContainer
}

type XMLBaseContainer struct {
	XMLName xml.Name `xml:"container"`
	XMLItem
	Color     string             `xml:"color,attr"`
	Conts     []XMLBaseContainer `xml:"container"`
	ListConts []XMLListContainer `xml:"listcontainer"`
	Labels    []XMLLabel         `xml:"label"`
	Textures  []XMLTexture       `xml:"texture"`
	Unicolor  []XMLUnicolor      `xml:"unicolor"`
}

type XMLListContainer struct {
	XMLName xml.Name `xml:"listcontainer"`
	XMLBaseContainer
}

type XMLLabel struct {
	XMLName xml.Name `xml:"label"`
	XMLItem
	TextSize string `xml:"textsize,attr"`
	VAlign   string `xml:"valign,attr"`
	HAlign   string `xml:"halign,attr"`
	FGColor  string `xml:"fgcolor,attr"`
	BGColor  string `xml:"bgcolor,attr"`
	Bold     string `xml:"bold,attr"`
	Text     string `xml:",chardata"`
}

type XMLTexture struct {
	XMLName xml.Name `xml:"texture"`
	XMLItem
	Texture   string `xml:",chardata"`
	ScaleDown string `xml:"scaledown,attr"`
}

type XMLUnicolor struct {
	XMLName xml.Name `xml:"unicolor"`
	XMLItem
	Color string `xml:",chardata"`
}

/*
###########################################
# Parser and window information
###########################################
*/

var bgcolor uint32
var bounds sdl.Rect

// gives back a XMLWIndow object which can be parsed further
func UnmarshalXMLFile(path string) (window XMLWindow, windowtype uint32, err error) {

	// open the file
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	// validate xml file.
	// it will be validated against
	// assets/style.xsd
	ok, err := validateXMLFile(file)
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

// get a drawable data.Container from an XMLWIndow
func (win *XMLWindow) Parse() (maincont *data.BaseContainer) {
	bgcolor = parseColor(win.Color)
	return win.parse(data.Vector{X: bounds.W, Y: bounds.H})
}

func validateXMLFile(file []byte) (valid bool, err error) {
	// validate xml with an xsd file
	xsdvalidate.Init()
	defer xsdvalidate.Cleanup()
	xsdhandler, err := xsdvalidate.NewXsdHandlerUrl("assets/style.xsd", xsdvalidate.ParsErrDefault)
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
			log.Println(err)
			log.Printf("Error in line: %d\n", err.(xsdvalidate.ValidationError).Errors[0].Line)
			log.Println(err.(xsdvalidate.ValidationError).Errors[0].Message)
			return false, nil
		default:
			return true, err
		}
	}

	return true, nil
}

/*
###########################################
# Struct Parser
###########################################
*/

// converts XMLContainer to data.Container
func (cont XMLBaseContainer) parse(psize data.Vector) (container *data.BaseContainer) {
	// get the size
	size := parseWH(cont.Width, cont.Height, psize)

	// the map of items any data.Container contains
	itemmap := make(map[string]data.Item)

	// Add the items to the list
	for it, item := range cont.Labels {
		itemmap["label"+strconv.Itoa(it)] = item.parse(size)
	}
	for it, item := range cont.Textures {
		itemmap["texture"+strconv.Itoa(it)] = item.parse(size)
	}
	for it, item := range cont.Unicolor {
		itemmap["unicolor"+strconv.Itoa(it)] = item.parse(size)
	}
	for it, item := range cont.Conts {
		itemmap["container"+strconv.Itoa(it)] = item.parse(size)
	}
	for it, item := range cont.ListConts {
		itemmap["listcontainer"+strconv.Itoa(it)] = item.parse(size)
	}

	// Construct container
	return &data.BaseContainer{
		Position: parseXY(cont.X, cont.Y, psize),
		Size:     size,
		BGcolor:  parseColor(cont.Color),
		Items:    itemmap,
		Events:   parseEvents(cont.OnEvent),
	}
}

func (cont XMLListContainer) parse(psize data.Vector) (listcontainer *data.ListContainer) {
	// get the size
	size := parseWH(cont.Width, cont.Height, psize)

	// the map of items any data.Container contains
	itemmap := make(map[string]data.ItemEntry)
	itemindex := 0

	// Add the items to the list
	for it, item := range cont.Labels {
		itemmap["label"+strconv.Itoa(it)] = data.ItemEntry{Index: itemindex, Item: item.parse(size)}
		itemindex++
	}
	for it, item := range cont.Textures {
		itemmap["texture"+strconv.Itoa(it)] = data.ItemEntry{Index: itemindex, Item: item.parse(size)}
		itemindex++
	}
	for it, item := range cont.Unicolor {
		itemmap["unicolor"+strconv.Itoa(it)] = data.ItemEntry{Index: itemindex, Item: item.parse(size)}
		itemindex++
	}
	for it, item := range cont.Conts {
		itemmap["container"+strconv.Itoa(it)] = data.ItemEntry{Index: itemindex, Item: item.parse(size)}
		itemindex++
	}
	for it, item := range cont.ListConts {
		itemmap["listcontainer"+strconv.Itoa(it)] = data.ItemEntry{Index: itemindex, Item: item.parse(size)}
		itemindex++
	}

	// Construct container
	return &data.ListContainer{
		Position: parseXY(cont.X, cont.Y, psize),
		Size:     size,
		BGcolor:  parseColor(cont.Color),
		Items:    itemmap,
		Events:   parseEvents(cont.OnEvent),
	}

}

// converts XMLUnicolor to data.Unicolor
func (uni XMLUnicolor) parse(psize data.Vector) (unicolor *data.Unicolor) {
	// Construct Unicolor
	return &data.Unicolor{
		Position: parseXY(uni.X, uni.Y, psize),
		Size:     parseWH(uni.Width, uni.Height, psize),
		Color:    parseColor(uni.Color),
		Events:   parseEvents(uni.OnEvent),
	}
}

// converts XMLLabel to data.Label
func (label XMLLabel) parse(psize data.Vector) (l *data.Label) {
	// Construct Label
	return &data.Label{
		Position: parseXY(label.X, label.Y, psize),
		Size:     parseWH(label.Width, label.Height, psize),
		Text:     parseText(label.Text),
		Textsize: parseInt(label.TextSize),
		Valign:   parseAlign(label.VAlign),
		Halign:   parseAlign(label.HAlign),
		Color:    parseColor(label.FGColor),
		BGcolor:  parseColor(label.BGColor),
		Bold:     parseBool(label.Bold),
		Events:   parseEvents(label.OnEvent),
	}
}

// converts XMLTexture to data.Texture
func (tex XMLTexture) parse(psize data.Vector) (texture *data.Texture) {
	// Construct Texture
	return &data.Texture{
		Position: parseXY(tex.X, tex.Y, psize),
		Size:     parseWH(tex.Width, tex.Height, psize),
		Texture:  parseImage(tex.Texture, parseXY(tex.Width, tex.Height, psize), parseBool(tex.ScaleDown)),
		Events:   parseEvents(tex.OnEvent),
	}
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
func parseColor(color string) (result uint32) {
	// return default if not specified
	if color == "" {
		return bgcolor
	}

	// remove whitespaces
	color = cleanString(color)

	// preprocess
	color = plug.PreParseString(color)

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
func parseBool(b string) (result bool) {
	b = cleanString(b)

	// preprocess
	b = plug.PreParseString(b)

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
func parseAlign(align string) (result data.Align) {
	align = cleanString(align)

	// preprocess
	align = plug.PreParseString(align)

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
func parseXY(x string, y string, parentSize data.Vector) (result data.Vector) {
	x = cleanString(x)
	y = cleanString(y)

	// preprocess
	x = plug.PreParseString(x)
	y = plug.PreParseString(y)

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
func parseWH(w string, h string, parentSize data.Vector) (result data.Vector) {
	result = parseXY(w, h, parentSize)

	if result.X == 0 {
		result.X = parentSize.X
	}
	if result.Y == 0 {
		result.Y = parentSize.Y
	}

	return
}

// parses a string to an int. Defaults to 0 if empty
func parseInt(integer string) (result int) {
	integer = cleanString(integer)

	// preprocess
	integer = plug.PreParseString(integer)

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
func parseImage(imagepath string, size data.Vector, downscale bool) (result *sdl.Surface) {
	imagepath = cleanString(imagepath)

	// preprocess
	imagepath = plug.PreParseString(imagepath)

	// return the already processed value of that image
	// if existent
	if val, ok := loadedimages[imagepath]; ok {
		return val
	}

	// get image object from file
	img, err := imgtools.ImageFromFile(imagepath)
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
	surf, err := imgtools.ImgToSurface(img)
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
		result, err = imgtools.ResizeSurface(surf, size.X, size.Y)
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
func parseText(text string) (result string) {
	text = cleanString(text)
	return plug.PreParseString(text)
}

// parses an ItemEvents struct to a map of events understood by the data package
// Entries may look like this:
// onevent="click:func1,rightclick:func2"
func parseEvents(onevent string) (result map[string]string) {
	result = make(map[string]string)

	onevent = cleanString(onevent)

	entries := strings.Split(onevent, ",")

	for _, entry := range entries {
		data := strings.Split(entry, ":")

		if len(data) == 2 {
			data[0] = parseText(data[0])
			data[1] = parseText(data[1])
			result[data[0]] = data[1]
		}
	}

	return
}

// cleans up whitespace before and after a string
func cleanString(s string) (cleaned string) {
	if s == "" {
		return s
	}

	// remove whitespaces before
	for strings.ContainsAny("\a\f\t\n\r\v", string(s[0])) {
		s = s[1:]
	}

	// remove whitespaces after
	for strings.ContainsAny("\a\f\t\n\r\v", string(s[len(s)-1])) {
		s = s[:len(s)-1]
	}

	return s
}
