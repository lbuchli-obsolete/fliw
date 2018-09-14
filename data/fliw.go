package data

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"

	"github.com/phoenixdevelops/fliw/io/imgtools"
)

/*
This is where all basic data types and their methods are.
*/

// TODO change all struct variables to uppercase

/*
##############################################################
# Section: Basic Types and functions
##############################################################
*/

type Vector struct {
	X int32
	Y int32
}

type FractionVector struct {
	X float32
	Y float32
}

// This wants to be an enum
type Align int

const (
	LEFT   Align = 0
	TOP    Align = 0
	CENTER Align = 1
	RIGHT  Align = 2
	BOTTOM Align = 2
)

// text sizes
const (
	TITLE     int = 128
	SUBTITLE  int = 64
	HEADER    int = 32
	SUBHEADER int = 24
	TEXT      int = 16
	SUBTEXT   int = 14
)

/*
###############################################################
# Section: Item & Container
###############################################################
*/

// Every type with a position and scale is considered an item.
type Item interface {
	Draw(*sdl.Surface) error
	GetPosition() Vector
	SetPosition(Vector)
	GetSize() Vector
	SetSize(Vector)
}

// This is the first (and the most important) item.
// It is used to group other items.
type Container struct {
	Position Vector
	Size     Vector
	BGcolor  uint32
	Items    map[string]Item
}

// Move the item to a pixel position
func (cont *Container) MoveItem(item string, pos Vector) {
	cont.Items[item].SetPosition(pos)
}

// Move the item to a fraction of the parent container size
func (cont *Container) MoveItemToFraction(item string, pos FractionVector) {
	cont.Items[item].SetPosition(Vector{int32(pos.X * float32(cont.Size.X)), int32(pos.Y * float32(cont.Size.Y))})
}

// Resize an Item to a specific pixel size
func (cont *Container) ResizeItem(item string, size Vector) {
	cont.Items[item].SetSize(size)
}

// Resize an Item to a fraction of the parent container size
func (cont *Container) ResizeItemToFraction(item string, size FractionVector) {
	cont.Items[item].SetSize(Vector{int32(size.X * float32(cont.Size.X)), int32(size.Y * float32(cont.Size.Y))})
}

// draw a container
// The container will let each item draw onto its own surface and then draw that onto the main surface
func (cont *Container) Draw(surf *sdl.Surface) (err error) {

	// let each item draw onto the surface
	for _, val := range cont.Items {
		isurface, err := sdl.CreateRGBSurface(0, val.GetSize().X, val.GetSize().Y, 32, 0, 0, 0, 0)
		if err != nil {
			return err
		}

		// also apply background color
		// flip bytes for sdl
		isurface.FillRect(nil, imgtools.UInt32ToColor(cont.BGcolor).Uint32())

		err = val.Draw(isurface)
		if err != nil {
			return err
		}

		pos := val.GetPosition()
		size := val.GetSize()

		// draw the item surface onto the container surface
		src_rect := sdl.Rect{X: 0, Y: 0, W: size.X, H: size.Y}
		dst_rect := sdl.Rect{X: pos.X, Y: pos.Y, W: pos.X + size.X, H: pos.Y + size.Y}
		isurface.Blit(&src_rect, surf, &dst_rect)

		isurface.Free()
	}

	return nil
}

// Add an item to the container
func (cont *Container) AddItem(name string, item Item) {
	cont.Items[name] = item
}

// Get an item from the container
func (cont *Container) GetItem(name string) (item Item) {
	return cont.Items[name]
}

// Getters and setters

func (cont *Container) GetPosition() (position Vector) {
	return cont.Position
}

func (cont *Container) SetPosition(position Vector) {
	cont.Position = position
}

func (cont *Container) GetSize() (size Vector) {
	return cont.Size
}

func (cont *Container) SetSize(size Vector) {
	cont.Size = size
}

/*
####################################################################
# Section: Basic item types
####################################################################
*/

/*
########################
# Subsection: Label
########################
*/

type Label struct {
	Position Vector
	Size     Vector
	Text     string
	Textsize int
	Valign   Align
	Halign   Align
	Color    uint32
	BGcolor  uint32
	Bold     bool
}

// Draw the item onto the parent surface
func (label *Label) Draw(surf *sdl.Surface) (err error) {

	var font *ttf.Font

	// load font
	if label.Bold {
		font, err = ttf.OpenFont("/usr/share/fonts/TTF/DejaVuSans-Bold.ttf", label.Textsize)
	} else {
		font, err = ttf.OpenFont("/usr/share/fonts/TTF/DejaVuSans.ttf", label.Textsize)
	}
	if err != nil {
		return err
	}
	defer font.Close()

	// Render text to surface
	text_surface, err := font.RenderUTF8Shaded(label.Text, imgtools.UInt32ToColor(label.Color), imgtools.UInt32ToColor(label.BGcolor))
	if err != nil {
		return err
	}
	defer text_surface.Free()

	// Calculate vertical and horizontal position on surface
	var coordinate_x int32
	var coordinate_y int32

	switch label.Halign {
	case LEFT:
		coordinate_x = 0
	case CENTER:
		coordinate_x = (int32(label.Size.X) - text_surface.W) / 2
	case RIGHT:
		coordinate_x = int32(label.Size.X) - text_surface.W
	}

	switch label.Valign {
	case TOP:
		coordinate_y = 0
	case CENTER:
		coordinate_y = (int32(label.Size.Y) - text_surface.H) / 2
	case BOTTOM:
		coordinate_y = int32(label.Size.Y) - text_surface.H
	}

	dst_rect := sdl.Rect{X: coordinate_x, Y: coordinate_y, W: coordinate_x + text_surface.W, H: coordinate_y + text_surface.H}

	// Color the background surface
	// convert to sdl color and back in order to make sure the color
	// is sdl compatible (no bytes flipped)
	surf.FillRect(nil, imgtools.UInt32ToColor(label.BGcolor).Uint32())

	// Draw onto final surface (Text aligned)
	text_surface.Blit(&sdl.Rect{X: 0, Y: 0, W: text_surface.W, H: text_surface.H}, surf, &dst_rect)

	return nil
}

// Getters and setters

func (label *Label) GetPosition() (position Vector) {
	return label.Position
}

func (label *Label) SetPosition(newposition Vector) {
	label.Position = newposition
}

func (label *Label) GetSize() (size Vector) {
	return label.Size
}

func (label *Label) SetSize(newsize Vector) {
	label.Size = newsize
}

/*
########################
# Subsection: Texture
########################
*/

type Texture struct {
	Position Vector
	Size     Vector
	Texture  *sdl.Surface
}

// Draw the item onto the parent surface
func (tex *Texture) Draw(surf *sdl.Surface) (err error) {
	src_rect := sdl.Rect{X: 0, Y: 0, W: tex.Size.X, H: tex.Size.Y}
	dst_rect := sdl.Rect{X: 0, Y: 0, W: tex.Size.X, H: tex.Size.Y}
	tex.Texture.Blit(&src_rect, surf, &dst_rect)

	return nil
}

// Getters and setters

func (tex *Texture) GetPosition() (position Vector) {
	return tex.Position
}

func (tex *Texture) SetPosition(position Vector) {
	tex.Position = position
}

func (tex *Texture) GetSize() (size Vector) {
	return tex.Size
}

func (tex *Texture) SetSize(size Vector) {
	tex.Size = size
}

/*
########################
# Subsection: Unicolor
########################
*/

type Unicolor struct {
	Position Vector
	Size     Vector
	Color    uint32
}

// Draw the item onto the parent surface
func (unic *Unicolor) Draw(surf *sdl.Surface) (err error) {
	rect := sdl.Rect{X: 0, Y: 0, W: unic.Size.X, H: unic.Size.Y}
	return surf.FillRect(&rect, unic.Color)
}

// Getters and setters

func (unic *Unicolor) GetPosition() (position Vector) {
	return unic.Position
}

func (unic *Unicolor) SetPosition(position Vector) {
	unic.Position = position
}

func (unic *Unicolor) GetSize() (size Vector) {
	return unic.Size
}

func (unic *Unicolor) SetSize(size Vector) {
	unic.Size = size
}
