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
	"github.com/phoenixdevelops/fliw/io/imgtools"
	"github.com/veandco/go-sdl2/sdl"
)

/*
parses your style.xml data
*/

type XMLItem interface {
	parse(data.Vector) data.Item
}

type XMLWindow struct {
	XMLName    xml.Name     `xml:"window"`
	WindowType string       `xml:"windowtype,attr"`
	BGColor    string       `xml:"bgcolor,attr"`
	Cont       XMLContainer `xml:"container"`
}

type XMLContainer struct {
	XMLName  xml.Name       `xml:"container"`
	X        string         `xml:"x,attr"`
	Y        string         `xml:"y,attr"`
	Width    string         `xml:"width,attr"`
	Height   string         `xml:"height,attr"`
	Color    string         `xml:"color,attr"`
	Conts    []XMLContainer `xml:"container"`
	Labels   []XMLLabel     `xml:"label"`
	Textures []XMLTexture   `xml:"texture"`
	Unicolor []XMLUnicolor  `xml:"unicolor"`
}

type XMLLabel struct {
	XMLName  xml.Name `xml:"label"`
	X        string   `xml:"x,attr"`
	Y        string   `xml:"y,attr"`
	Width    string   `xml:"width,attr"`
	Height   string   `xml:"height,attr"`
	TextSize string   `xml:"textsize,attr"`
	VAlign   string   `xml:"valign,attr"`
	HAlign   string   `xml:"halign,attr"`
	FGColor  string   `xml:"fgcolor,attr"`
	BGColor  string   `xml:"bgcolor,attr"`
	Bold     string   `xml:"bold,attr"`
	Text     string   `xml:",chardata"`
}

type XMLTexture struct {
	XMLName   xml.Name `xml:"texture"`
	X         string   `xml:"x,attr"`
	Y         string   `xml:"y,attr"`
	Width     string   `xml:"width,attr"`
	Height    string   `xml:"height,attr"`
	Texture   string   `xml:",chardata"`
	ScaleDown string   `xml:"scaledown,attr"`
}

type XMLUnicolor struct {
	XMLName xml.Name `xml:"unicolor"`
	X       string   `xml:"x,attr"`
	Y       string   `xml:"y,attr"`
	Width   string   `xml:"width,attr"`
	Height  string   `xml:"height,attr"`
	Color   string   `xml:",chardata"`
}

/*
###########################################
# Parser and window information
###########################################
*/

var bgcolor uint32

// parses XML data to drawable data.* data
func ParseXMLFile(path string) (maincont *data.Container, backgroundcolor uint32, windowtype string, err error) {
	// open the file
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return maincont, backgroundcolor, windowtype, err
	}

	win := XMLWindow{
		BGColor:    "0x10171e",
		WindowType: "popup_menu",
	}

	// unmarshal the file
	err = xml.Unmarshal([]byte(file), &win)
	if err != nil {
		return maincont, backgroundcolor, windowtype, err
	}

	// get the display size
	bounds, err := sdl.GetDisplayBounds(0)
	if err != nil {
		return maincont, backgroundcolor, windowtype, err
	}

	bgcolor = parseColor(win.BGColor)

	return win.Cont.parse(data.Vector{X: bounds.W, Y: bounds.H}), bgcolor, win.WindowType, nil
}

/*
###########################################
# Struct Parser
###########################################
*/

// converts XMLContainer to data.Container
func (cont XMLContainer) parse(psize data.Vector) (container *data.Container) {
	// get the size
	size := parseXY(cont.Width, cont.Height, psize)

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

	// Construct container
	return &data.Container{
		Position: parseXY(cont.X, cont.Y, psize),
		Size:     size,
		BGcolor:  parseColor(cont.Color),
		Items:    itemmap,
	}
}

// converts XMLUnicolor to data.Unicolor
func (uni XMLUnicolor) parse(psize data.Vector) (unicolor *data.Unicolor) {
	// Construct Unicolor
	return &data.Unicolor{
		Position: parseXY(uni.X, uni.Y, psize),
		Size:     parseXY(uni.Width, uni.Height, psize),
		Color:    parseColor(uni.Color),
	}
}

// converts XMLLabel to data.Label
func (label XMLLabel) parse(psize data.Vector) (l *data.Label) {
	// Construct Label
	return &data.Label{
		Position: parseXY(label.X, label.Y, psize),
		Size:     parseXY(label.Width, label.Height, psize),
		Text:     parseText(label.Text),
		Textsize: parseInt(label.TextSize),
		Valign:   parseAlign(label.VAlign),
		Halign:   parseAlign(label.HAlign),
		Color:    parseColor(label.FGColor),
		BGcolor:  parseColor(label.BGColor),
		Bold:     parseBool(label.Bold),
	}
}

// converts XMLTexture to data.Texture
func (tex XMLTexture) parse(psize data.Vector) (texture *data.Texture) {
	// Construct Texture
	return &data.Texture{
		Position: parseXY(tex.X, tex.Y, psize),
		Size:     parseXY(tex.Width, tex.Height, psize),
		Texture:  parseImage(tex.Texture, parseXY(tex.Width, tex.Height, psize), parseBool(tex.ScaleDown)),
	}
}

/*
###########################################
# String Parser
###########################################
*/

// parses a string containing a hex color to a uint32
// defaults to bgcolor
// representing said color
func parseColor(color string) (result uint32) {

	// return default if not specified
	if color == "" {
		return bgcolor
	}

	// remove whitespaces
	color = cleanString(color)

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

	// return the byte array as uint32
	return binary.LittleEndian.Uint32(val)
}

// parses a string to a bool value. Defaults to false if string is empty
func parseBool(b string) (result bool) {
	b = cleanString(b)

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

// parses x and y strings to a data.Vector size / position. Defaults to 0 if string is empty.
// Uses parentSize for percentual interpretation
func parseXY(x string, y string, parentSize data.Vector) (result data.Vector) {
	x = cleanString(x)
	y = cleanString(y)

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

// parses a string to an int. Defaults to 0 if empty
func parseInt(integer string) (result int) {
	integer = cleanString(integer)

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

// parses a string (path) to an *sdl.Surface. Defaults to a unicolored surface if empty
func parseImage(imagepath string, size data.Vector, downscale bool) (result *sdl.Surface) {
	imagepath = cleanString(imagepath)

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

	return result
}

// parses a string to a string (removes whitespace before and after)
func parseText(text string) (result string) {
	return cleanString(text)
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
