package data

import (
	"github.com/phoenixdevelops/fliw/image"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

/*
This is where all basic data types and their methods are.
*/

/*
##############################################################
# Section: Basic Types and functions
##############################################################
*/

// Vector is a struct holding an X and Y value.
// It can be used for position, size, ...
type Vector struct {
	X int32
	Y int32
}

// FractionVector is a struct holding the fraction of
// an X and Y value. Needs to be recalculated before use as
// position or size
type FractionVector struct {
	X float32
	Y float32
}

// Align wants to be an enum
type Align int

// constants of the wanna-be enum Align
const (
	LEFT   Align = 0
	TOP    Align = 0
	CENTER Align = 1
	RIGHT  Align = 2
	BOTTOM Align = 2
)

// unused function that will fail to compile
// if any of the listed structs are not in the interface
func checkIntefraceSatisfaction() {
	var _ Container = (*BaseContainer)(nil)
	var _ Container = (*ListContainer)(nil)
	var _ Item = (*Label)(nil)
	var _ Item = (*Texture)(nil)
	var _ Item = (*Unicolor)(nil)
}

/*
###############################################################
# Section: Item & Container
###############################################################
*/

// Item is the base interface of all elements
type Item interface {
	GetUID() uint
	Draw(*sdl.Surface) error
	GetPosition() Vector
	SetPosition(Vector)
	GetSize() Vector
	SetSize(Vector)
	GetEvent(string) string
	GetEvents() map[string]string
}

// Container is an item containing other items.
type Container interface {
	Item
	MoveItem(int, Vector)
	MoveItemToFraction(int, FractionVector)
	ResizeItem(int, Vector)
	ResizeItemToFraction(int, FractionVector)
	AddItem(Item)
	GetItem(int) Item
	GetItems() []Item
	GetItemAt(Vector) Item
	GetIsLink() bool
	SetLink(bool)
}

type ItemBase struct {
	UID      uint
	Position Vector
	Size     Vector
	Events   map[string]string
}

// GetUID returns the unique identifier of the item
func (base *ItemBase) GetUID() (uid uint) {
	return base.UID
}

// GetPosition returns the position
func (base *ItemBase) GetPosition() (position Vector) {
	return base.Position
}

// SetPosition sets the position
func (base *ItemBase) SetPosition(position Vector) {
	base.Position = position
}

// GetSize returns the size
func (base *ItemBase) GetSize() (size Vector) {
	return base.Size
}

// SetSize sets the size
func (base *ItemBase) SetSize(size Vector) {
	base.Size = size
}

// GetEvent returns the name of the function to call
// when event is invoked
func (base *ItemBase) GetEvent(event string) (f string) {
	return base.Events[event]
}

// GetEvents returns a map of all event assignments
func (base *ItemBase) GetEvents() map[string]string {
	return base.Events
}

// ContainerBase is the base for every container struct
type ContainerBase struct {
	ItemBase
	BGcolor uint32
	Items   []Item
	IsLink  bool
}

// MoveItem moves the item to a pixel position
func (cont *ContainerBase) MoveItem(index int, pos Vector) {
	cont.Items[index].SetPosition(pos)
}

// MoveItemToFraction moves the item to a fraction of the parent container size
func (cont *ContainerBase) MoveItemToFraction(index int, pos FractionVector) {
	cont.Items[index].SetPosition(Vector{int32(pos.X * float32(cont.Size.X)), int32(pos.Y * float32(cont.Size.Y))})
}

// ResizeItem resizes an Item to a specific pixel size
func (cont *ContainerBase) ResizeItem(index int, size Vector) {
	cont.Items[index].SetSize(size)
}

// ResizeItemToFraction resizes an Item to a fraction of the parent container size
func (cont *ContainerBase) ResizeItemToFraction(index int, size FractionVector) {
	cont.Items[index].SetSize(Vector{int32(size.X * float32(cont.Size.X)), int32(size.Y * float32(cont.Size.Y))})
}

// AddItem adds an item to the container
func (cont *ContainerBase) AddItem(item Item) {
	cont.Items = append(cont.Items, item)
}

// GetItem gets an item from the container
func (cont *ContainerBase) GetItem(index int) (item Item) {
	return cont.Items[index]
}

// GetItems gets all child items of a container
func (cont *ContainerBase) GetItems() []Item {
	return cont.Items
}

// GetIsLink tells wether this container is used as a link to another XML file
func (cont *ContainerBase) GetIsLink() bool {
	return cont.IsLink
}

// SetLink sets wether this is used as a link
func (cont *ContainerBase) SetLink(isLink bool) {
	cont.IsLink = isLink
}

/*
####################################################################
# Section: Basic item types
####################################################################
*/

/*
########################
# Subsection: BaseContainer
########################
*/

// BaseContainer is the first (and the most important) item.
// It is used to group other items.
type BaseContainer struct {
	ContainerBase
}

// Draw draws a container onto a surface
// The container will let each item draw onto its own surface and then draw that onto the main surface
func (cont *BaseContainer) Draw(surf *sdl.Surface) (err error) {
	// let each item draw onto the surface
	for _, val := range cont.Items {
		// get size and position of the item
		pos := val.GetPosition()
		size := val.GetSize()

		isurface, err := sdl.CreateRGBSurface(0, size.X, size.Y, 32, 0, 0, 0, 0)
		if err != nil {
			return err
		}

		// also apply background color
		// flip bytes for sdl
		isurface.FillRect(nil, image.UInt32ToColor(cont.BGcolor).Uint32())

		err = val.Draw(isurface)
		if err != nil {
			return err
		}

		// draw the item surface onto the container surface
		srcRect := sdl.Rect{X: 0, Y: 0, W: size.X, H: size.Y}
		dstRect := sdl.Rect{X: pos.X, Y: pos.Y, W: size.X, H: size.Y}
		isurface.Blit(&srcRect, surf, &dstRect)

		isurface.Free()
	}

	return
}

// GetItemAt gets you the item at position pos
func (cont *BaseContainer) GetItemAt(pos Vector) Item {
	for _, item := range cont.Items {
		position := item.GetPosition()
		size := item.GetSize()

		// Check if pos is inside item
		if position.X <= pos.X && position.Y <= pos.Y {
			if (position.X+size.X) > pos.X && (position.Y+size.Y) > pos.Y {
				return item
			}
		}
	}

	// if nothing was found, the area must be the container itself
	return cont
}

/*
########################
# Subsection: ListContainer
########################
*/

// ListContainer groups, just like a normal container,
// items, but unlike the normal container
// it wont stack them ontop of each other but list them
// below each other
type ListContainer struct {
	ContainerBase
}

// Draw draws a listcontainer onto a surface
// The container will let each item draw onto its own surface and then draw that onto the main surface
// in a listcontainer all items are drawn below each other with item pos y as offset
func (cont *ListContainer) Draw(surf *sdl.Surface) (err error) {

	yoffset := int32(0)

	// let each item draw onto the surface
	for _, item := range cont.Items {
		pos := item.GetPosition()
		size := item.GetSize()

		// if not in picure don't draw
		if yoffset+pos.Y > surf.H {
			continue
		}

		isurface, err := sdl.CreateRGBSurface(0, size.X, size.Y, 32, 0, 0, 0, 0)
		if err != nil {
			return err
		}

		// also apply background color
		// flip bytes for sdl
		isurface.FillRect(nil, image.UInt32ToColor(cont.BGcolor).Uint32())

		err = item.Draw(isurface)
		if err != nil {
			return err
		}

		// draw the item surface onto the container surface
		srcRect := sdl.Rect{X: 0, Y: 0, W: size.X, H: size.Y}
		dstRect := sdl.Rect{X: pos.X, Y: pos.Y + yoffset, W: size.X, H: size.Y}
		isurface.Blit(&srcRect, surf, &dstRect)

		isurface.Free()

		yoffset += pos.Y + size.Y
	}

	return nil
}

// GetItemAt gets you the item at position pos
func (cont *ListContainer) GetItemAt(pos Vector) Item {
	var yoffset int32

	for _, item := range cont.Items {
		position := item.GetPosition()
		size := item.GetSize()

		// Check if pos is inside item
		if position.X <= pos.X && position.Y+yoffset <= pos.Y {
			if (position.X+size.X) > pos.X && (position.Y+size.Y+yoffset) > pos.Y {
				return item
			}
		}

		yoffset += position.Y + size.Y
	}

	// if nothing was found, the area must be the container itself
	return cont
}

/*
########################
# Subsection: Label
########################
*/

// Label is an item that holds basic, short text
type Label struct {
	ItemBase

	Text     string
	Textsize int
	Valign   Align
	Halign   Align
	Color    uint32
	BGcolor  uint32
	Bold     bool
}

// Draw draws the item onto the parent surface
func (label *Label) Draw(surf *sdl.Surface) (err error) {

	// Color the background surface
	// convert to sdl color and back in order to make sure the color
	// is sdl compatible (no bytes flipped)
	surf.FillRect(nil, image.UInt32ToColor(label.BGcolor).Uint32())

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

	// if text is not empty
	if label.Text != "" {

		// Render text to surface
		textSurface, err := font.RenderUTF8Shaded(label.Text, image.UInt32ToColor(label.Color), image.UInt32ToColor(label.BGcolor))
		if err != nil {
			return err
		}
		defer textSurface.Free()

		// Calculate vertical and horizontal position on surface
		var coordinateX int32
		var coordinateY int32

		switch label.Halign {
		case LEFT:
			coordinateX = 0
		case CENTER:
			coordinateX = (int32(label.Size.X) - textSurface.W) / 2
		case RIGHT:
			coordinateX = int32(label.Size.X) - textSurface.W
		}

		switch label.Valign {
		case TOP:
			coordinateY = 0
		case CENTER:
			coordinateY = (int32(label.Size.Y) - textSurface.H) / 2
		case BOTTOM:
			coordinateY = int32(label.Size.Y) - textSurface.H
		}

		dstRect := sdl.Rect{X: coordinateX, Y: coordinateY, W: coordinateX + textSurface.W, H: coordinateY + textSurface.H}

		// Draw onto final surface (Text aligned)
		textSurface.Blit(&sdl.Rect{X: 0, Y: 0, W: textSurface.W, H: textSurface.H}, surf, &dstRect)

	}

	return
}

/*
########################
# Subsection: Texture
########################
*/

// Texture is an item containing a texture/picture
type Texture struct {
	ItemBase

	Texture *sdl.Surface
}

// Draw the item onto the parent surface
func (tex *Texture) Draw(surf *sdl.Surface) (err error) {
	srcRect := sdl.Rect{X: 0, Y: 0, W: tex.Size.X, H: tex.Size.Y}
	dstRect := sdl.Rect{X: 0, Y: 0, W: tex.Size.X, H: tex.Size.Y}
	tex.Texture.Blit(&srcRect, surf, &dstRect)

	return nil
}

/*
########################
# Subsection: Unicolor
########################
*/

// Unicolor is an item with just one color
type Unicolor struct {
	ItemBase

	Color uint32
}

// Draw the item onto the parent surface
func (unic *Unicolor) Draw(surf *sdl.Surface) (err error) {
	rect := sdl.Rect{X: 0, Y: 0, W: unic.Size.X, H: unic.Size.Y}
	return surf.FillRect(&rect, unic.Color)
}
