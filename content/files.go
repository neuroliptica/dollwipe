// files.go: definitions for manipulating files and it's content.

package content

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"time"
)

// General image decoder and encoder function types.
type (
	Decoder = func(io.Reader) (image.Image, error)
	Encoder = func(io.Writer, image.Image) error
)

// Colorize image represented as a slice of bytes.
func colorizeBytes(img []byte, decoder Decoder) (image.Image, error) {
	r := bytes.NewReader(img)
	modified, err := colorize(r, decoder)
	if err != nil {
		return nil, err
	}
	return modified, nil
}

// General colorization function.
// Won't modify passed []byte slice. Instead will create a new one.
func GeneralColorize(img []byte, decoder Decoder, encoder Encoder) ([]byte, error) {
	colorized, err := colorizeBytes(img, decoder)
	if err != nil {
		return nil, err
	}
	colorizedBuf := new(bytes.Buffer)
	err = encoder(colorizedBuf, colorized)
	if err != nil {
		return nil, err
	}
	return colorizedBuf.Bytes(), nil
}

// Colorize jpg/jpeg media represented in bytes.
func JpegColorize(img []byte) ([]byte, error) {
	return GeneralColorize(img, jpeg.Decode, func(r io.Writer, c image.Image) error {
		return jpeg.Encode(r, c, nil)
	})
}

// Colorize png media represented in bytes.
func PngColorize(img []byte) ([]byte, error) {
	return GeneralColorize(img, png.Decode, png.Encode)
}

// Will create random RGBA layer and paint it over passed image.
func colorize(r io.Reader, decoder Decoder) (image.Image, error) {
	rand.Seed(time.Now().UnixNano())
	img, err := decoder(r)
	if err != nil {
		return nil, err
	}
	size := img.Bounds().Size()

	mutable := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))

	ar, ag, ab, aa :=
		float64(rand.Uint32()%0x4000)/0xFFFF,
		float64(rand.Uint32()%0x4000)/0xFFFF,
		float64(rand.Uint32()%0x4000)/0xFFFF,
		float64(rand.Uint32()%0x4000)/0xFFFF

	for x := 0; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			r, g, b, a := img.At(x, y).RGBA()
			br := float64(r) / 0xFFFF
			bg := float64(g) / 0xFFFF
			bb := float64(b) / 0xFFFF
			ba := float64(a) / 0xFFFF

			oa := aa + ba*(1-aa)
			or := ar + br*(1-aa)
			og := ag + bg*(1-aa)
			ob := ab + bb*(1-aa)

			color := color.RGBA64{
				R: uint16(0xFFFF * or * oa),
				G: uint16(0xFFFF * og * oa),
				B: uint16(0xFFFF * ob * oa),
				A: uint16(0xFFFF * oa),
			}

			mutable.Set(x, y, color)
		}
	}
	return mutable, nil
}
