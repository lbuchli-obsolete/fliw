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
func checkSatisfaction() {
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
	MoveItem(string, Vector)
	MoveItemToFraction(string, FractionVector)
	ResizeItem(string, Vector)
	ResizeItemToFraction(string, FractionVector)
	AddItem(string, Item)
	GetItem(string) Item
	GetItems() map[string]Item
	GetItemAt(Vector) Item
	GetIsLink() bool
	SetLink(bool)
}

// BaseContainer is the first (and the most important) item.
// It is used to group other items.
type BaseContainer struct {
	Position Vector
	Size     Vector
	Events   map[string]string

	BGcolor uint32
	Items   map[string]Item
	IsLink  bool
}

// MoveItem moves the item to a pixel position
func (cont *BaseContainer) MoveItem(item string, pos Vector) {
	cont.Items[item].SetPosition(pos)
}

// MoveItemToFraction moves the item to a fraction of the parent container size
func (cont *BaseContainer) MoveItemToFraction(item string, pos FractionVector) {
	cont.Items[item].SetPosition(Vector{int32(pos.X * float32(cont.Size.X)), int32(pos.Y * float32(cont.Size.Y))})
}

// ResizeItem resizes an Item to a specific pixel size
func (cont *BaseContainer) ResizeItem(item string, size Vector) {
	cont.Items[item].SetSize(size)
}

// ResizeItemToFraction resizes an Item to a fraction of the parent container size
func (cont *BaseContainer) ResizeItemToFraction(item string, size FractionVector) {
	cont.Items[item].SetSize(Vector{int32(size.X * float32(cont.Size.X)), int32(size.Y * float32(cont.Size.Y))})
}

// Draw draws a container onto a surface
// The container will let each item draw onto its own surface and then draw that onto the main surface
func (cont *BaseContainer) Draw(surf *sdl.Surface) (err error) {

	// let each item draw onto the surface
	for _, val := range cont.Items {
		isurface, err := sdl.CreateRGBSurface(0, val.GetSize().X, val.GetSize().Y, 32, 0, 0, 0, 0)
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

		pos := val.GetPosition()
		size := val.GetSize()

		// draw the item surface onto the container surface
		srcRect := sdl.Rect{X: 0, Y: 0, W: size.X, H: size.Y}
		dstRect := sdl.Rect{X: pos.X, Y: pos.Y, W: size.X, H: size.Y}
		isurface.Blit(&srcRect, surf, &dstRect)

		isurface.Free()
	}

	return nil
}

// AddItem adds an item to the container
func (cont *BaseContainer) AddItem(name string, item Item) {
	cont.Items[name] = item
}

// GetItem gets an item from the container
func (cont *BaseContainer) GetItem(name string) (item Item) {
	return cont.Items[name]
}

// GetItems gets all child items of a container
func (cont *BaseContainer) GetItems() map[string]Item {
	return cont.Items
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

// Getters and setters

// GetPosition returns the position
func (cont *BaseContainer) GetPosition() (position Vector) {
	return cont.Position
}

// SetPosition sets the position
func (cont *BaseContainer) SetPosition(position Vector) {
	cont.Position = position
}

// GetSize returns the size
func (cont *BaseContainer) GetSize() (size Vector) {
	return cont.Size
}

// SetSize sets the size
func (cont *BaseContainer) SetSize(size Vector) {
	cont.Size = size
}

// GetEvent returns the name of the function to call
// when event is invoked
func (cont *BaseContainer) GetEvent(event string) (f string) {
	return cont.Events[event]
}

// GetEvents returns a map of all event assignments
func (cont *BaseContainer) GetEvents() map[string]string {
	return cont.Events
}

// GetIsLink tells wether this container is used as a link to another XML file
func (cont *BaseContainer) GetIsLink() bool {
	return cont.IsLink
}

// SetLink sets wether this is used as a link
func (cont *BaseContainer) SetLink(isLink bool) {
	cont.IsLink = isLink
}

/*
####################################################################
# Section: Basic item types
####################################################################
*/

/*
########################
# Subsection: ListContainer
########################
*/

// ItemEntry is an entry to a ListContainer item map
type ItemEntry struct {
	Index int
	Item  Item
}

// ListContainer groups, just like a normal container,
// items, but unlike the normal container
// it wont stack them ontop of each other but list them
// below each other
type ListContainer struct {
	Position Vector
	Size     Vector
	Events   map[string]string

	BGcolor uint32
	Items   map[string]ItemEntry
	IsLink  bool
}

// MoveItem moves the item to a pixel position
// This will determine its margin
func (cont *ListContainer) MoveItem(item string, pos Vector) {
	cont.Items[item].Item.SetPosition(pos)
}

// MoveItemToFraction moves the item to a fraction of the parent container size
// This will determine its margin
func (cont *ListContainer) MoveItemToFraction(item string, pos FractionVector) {
	cont.Items[item].Item.SetPosition(Vector{int32(pos.X * float32(cont.Size.X)), int32(pos.Y * float32(cont.Size.Y))})
}

// ResizeItem resizes an Item to a specific pixel size
func (cont *ListContainer) ResizeItem(item string, size Vector) {
	cont.Items[item].Item.SetSize(size)
}

// ResizeItemToFraction resizes an Item to a fraction of the parent container size
func (cont *ListContainer) ResizeItemToFraction(item string, size FractionVector) {
	cont.Items[item].Item.SetSize(Vector{int32(size.X * float32(cont.Size.X)), int32(size.Y * float32(cont.Size.Y))})
}

func (cont *ListContainer) getSortedItems() (items []Item) {
	items = make([]Item, len(cont.Items))

	// Put items in correct order
	for _, val := range cont.Items {
		items[val.Index] = val.Item
	}

	return
}

// Draw draws a listcontainer onto a surface
// The container will let each item draw onto its own surface and then draw that onto the main surface
// in a listcontainer all items are drawn below each other with item pos y as offset
func (cont *ListContainer) Draw(surf *sdl.Surface) (err error) {

	yoffset := int32(0)

	// let each item draw onto the surface
	for _, item := range cont.getSortedItems() {
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

// AddItem adds an item to the container
func (cont *ListContainer) AddItem(name string, item Item) {
	// index is one bigger than the last one
	cont.Items[name] = ItemEntry{len(cont.Items), item}
}

// GetItem gets an item from the container
func (cont *ListContainer) GetItem(name string) (item Item) {
	return cont.Items[name].Item
}

// GetItems gets all child items of a container
func (cont *ListContainer) GetItems() (items map[string]Item) {
	for key, val := range cont.Items {
		items[key] = val.Item
	}
	return
}

// GetItemAt gets you the item at position pos
func (cont *ListContainer) GetItemAt(pos Vector) Item {
	var yoffset int32

	for _, item := range cont.getSortedItems() {
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

// Getters and setters

// GetPosition returns the position
func (cont *ListContainer) GetPosition() (position Vector) {
	return cont.Position
}

// SetPosition sets the position
func (cont *ListContainer) SetPosition(position Vector) {
	cont.Position = position
}

// GetSize returns the size
func (cont *ListContainer) GetSize() (size Vector) {
	return cont.Size
}

// SetSize sets the size
func (cont *ListContainer) SetSize(size Vector) {
	cont.Size = size
}

// GetEvent gets the name of the function to call when event is invoked
func (cont *ListContainer) GetEvent(event string) (f string) {
	return cont.Events[event]
}

// GetEvents returns a map of event assignments
func (cont *ListContainer) GetEvents() map[string]string {
	return cont.Events
}

// GetIsLink returns wether the container is used as a link to another XML file
func (cont *ListContainer) GetIsLink() bool {
	return cont.IsLink
}

// SetLink sets wether this is used as a link
func (cont *ListContainer) SetLink(isLink bool) {
	cont.IsLink = isLink
}

/*
########################
# Subsection: Label
########################
*/

// Label is an item that holds basic, short text
type Label struct {
	Position Vector
	Size     Vector
	Events   map[string]string

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

// Getters and setters

// GetPosition returns the position
func (label *Label) GetPosition() (position Vector) {
	return label.Position
}

// SetPosition sets the position
func (label *Label) SetPosition(newposition Vector) {
	label.Position = newposition
}

// GetSize returns the size
func (label *Label) GetSize() (size Vector) {
	return label.Size
}

// SetSize sets the size
func (label *Label) SetSize(newsize Vector) {
	label.Size = newsize
}

// GetEvent gets the name of the function to call when event is invoked
func (label *Label) GetEvent(event string) (f string) {
	return label.Events[event]
}

// GetEvents returns a map of event assignments
func (label *Label) GetEvents() map[string]string {
	return label.Events
}

/*
########################
# Subsection: Texture
########################
*/

// Texture is an item containing a texture/picture
type Texture struct {
	Position Vector
	Size     Vector
	Events   map[string]string

	Texture *sdl.Surface
}

// Draw the item onto the parent surface
func (tex *Texture) Draw(surf *sdl.Surface) (err error) {
	srcRect := sdl.Rect{X: 0, Y: 0, W: tex.Size.X, H: tex.Size.Y}
	dstRect := sdl.Rect{X: 0, Y: 0, W: tex.Size.X, H: tex.Size.Y}
	tex.Texture.Blit(&srcRect, surf, &dstRect)

	return nil
}

// Getters and setters

// GetPosition returns the position
func (tex *Texture) GetPosition() (position Vector) {
	return tex.Position
}

// SetPosition sets the position
func (tex *Texture) SetPosition(position Vector) {
	tex.Position = position
}

// GetSize returns the size
func (tex *Texture) GetSize() (size Vector) {
	return tex.Size
}

// SetSize sets the size
func (tex *Texture) SetSize(size Vector) {
	tex.Size = size
}

// GetEvent returns the name of the function to call when event is invoked
func (tex *Texture) GetEvent(event string) (f string) {
	return tex.Events[event]
}

// GetEvents returns a list of event assignments
func (tex *Texture) GetEvents() map[string]string {
	return tex.Events
}

/*
########################
# Subsection: Unicolor
########################
*/

// Unicolor is an item with just one color
type Unicolor struct {
	Position Vector
	Size     Vector
	Events   map[string]string

	Color uint32
}

// Draw the item onto the parent surface
func (unic *Unicolor) Draw(surf *sdl.Surface) (err error) {
	rect := sdl.Rect{X: 0, Y: 0, W: unic.Size.X, H: unic.Size.Y}
	return surf.FillRect(&rect, unic.Color)
}

// Getters and setters

// GetPosition returns the position
func (unic *Unicolor) GetPosition() (position Vector) {
	return unic.Position
}

// SetPosition sets the position
func (unic *Unicolor) SetPosition(position Vector) {
	unic.Position = position
}

// GetSize returns the size
func (unic *Unicolor) GetSize() (size Vector) {
	return unic.Size
}

// SetSize sets the size
func (unic *Unicolor) SetSize(size Vector) {
	unic.Size = size
}

// GetEvent gets the name of the function to call when event is invoked
func (unic *Unicolor) GetEvent(event string) (f string) {
	return unic.Events[event]
}

// GetEvents gets a map of event assignments
func (unic *Unicolor) GetEvents() map[string]string {
	return unic.Events
}
