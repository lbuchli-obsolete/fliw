package imgtools

import (
	"encoding/binary"
	"image"
	"os"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"

	_ "image/jpeg"
	_ "image/png"
)

/*
System icon & image tools
*/

// Turn a image (image.Image) into a sdl.Surface
func ImgToSurface(img image.Image) (surface *sdl.Surface, err error) {
	// Credit to https://github.com/veandco/go-sdl2/issues/116#issuecomment-96056082
	rgba := image.NewRGBA(img.Bounds())
	w, h := img.Bounds().Max.X, img.Bounds().Max.Y
	s, err := sdl.CreateRGBSurface(0, int32(w), int32(h), 32, 0, 0, 0, 0)
	if err != nil {
		return s, err
	}
	rgba.Pix = s.Pixels()

	for y := 0; y < h; y += 1 {
		for x := 0; x < w; x += 1 {
			c := img.At(x, y)
			rgba.Set(x, y, c)
		}
	}

	return s, nil
}

// Downscale a surface to a given size
func ResizeSurface(surf *sdl.Surface, newx int32, newy int32) (resizedsurf *sdl.Surface, err error) {
	resizedsurf, err = sdl.CreateRGBSurface(0, newx, newy, 32, 0, 0, 0, 0)
	if err != nil {
		return resizedsurf, err
	}

	rgba := image.NewRGBA(image.Rect(0, 0, int(newx), int(newy)))
	rgba.Pix = resizedsurf.Pixels()

	pixels := surf.Pixels()

	// the factor from newscale to oldscale
	scalex := float32(surf.W) / float32(newx)
	scaley := float32(surf.H) / float32(newy)

	for y := int32(0); y < newy; y += 1 {
		for x := int32(0); x < newx; x += 1 {
			// the byte index of the color currently looking at
			index := int32((float32(surf.BytesPerPixel()) * (float32(surf.W*y) * scaley)) +
				(float32(int32(surf.BytesPerPixel())*x) * scalex))

			// set the pixel using magic
			rgba.Set(int(x), int(y), UInt32ToColor(binary.BigEndian.Uint32(pixels[index:index+4])))
		}
	}

	return resizedsurf, err
}

// Load an image from an existing file
func ImageFromFile(path string) (img image.Image, err error) {
	file, err := os.Open(path)
	if err != nil {
		return img, err
	}
	defer file.Close()

	img, _, err = image.Decode(file)

	return img, err
}

func UInt32ToColor(ui uint32) (color sdl.Color) {
	bytes := (*[4]byte)(unsafe.Pointer(&ui))[:]
	return sdl.Color{R: bytes[0], G: bytes[1], B: bytes[2], A: bytes[3]}
}
