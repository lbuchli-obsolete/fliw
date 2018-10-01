package image

import (
	"image"
	"image/color"
	"os"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"

	// is needed in order to support these file types
	_ "image/jpeg"
	_ "image/png"
)

/*
System icon & image tools
*/

// ImgToSurface turns an image (image.Image) into a sdl.Surface
func ImgToSurface(img image.Image) (surface *sdl.Surface, err error) {
	// Credit to https://github.com/veandco/go-sdl2/issues/116#issuecomment-96056082
	rgba := image.NewRGBA(img.Bounds())
	w, h := img.Bounds().Max.X, img.Bounds().Max.Y
	s, err := sdl.CreateRGBSurface(0, int32(w), int32(h), 32, 0, 0, 0, 0)
	if err != nil {
		return s, err
	}
	rgba.Pix = s.Pixels()

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			rgba.Set(x, y, color.RGBA{uint8(b), uint8(g), uint8(r), uint8(a)})
		}
	}

	return s, nil
}

// ResizeSurface scales a surface down to a given size
func ResizeSurface(surf *sdl.Surface, newx int32, newy int32) (resizedsurf *sdl.Surface, err error) {
	var scale float64

	if (surf.W - newx) > (surf.H - newy) {
		scale = float64(newx) / float64(surf.W)
	} else {
		scale = float64(newy) / float64(surf.H)
	}

	resizedsurf, err = sdl.CreateRGBSurface(0, int32(float64(surf.W)*scale), int32(float64(surf.H)*scale), 32, 0, 0, 0, 0)
	if err != nil {
		return resizedsurf, err
	}

	surf.BlitScaled(&sdl.Rect{X: 0, Y: 0, W: surf.W, H: surf.H}, resizedsurf, &sdl.Rect{X: 0, Y: 0, W: resizedsurf.W, H: resizedsurf.H})

	return
}

// FromFile loads an image from an existing file
func FromFile(path string) (img image.Image, err error) {
	file, err := os.Open(path)
	if err != nil {
		return img, err
	}
	defer file.Close()

	img, _, err = image.Decode(file)

	return img, err
}

// UInt32ToColor turns an uint32 to a sdl.Color
func UInt32ToColor(ui uint32) (color sdl.Color) {
	bytes := (*[4]byte)(unsafe.Pointer(&ui))[:]
	return sdl.Color{R: bytes[0], G: bytes[1], B: bytes[2], A: bytes[3]}
}
