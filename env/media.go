// media.go: working with media: images, videos, etc.

package env

import (
	"dollwipe/content"
	"math/rand"
	"time"
)

// General file type.
// All methods require *File, but will never modify the original instance.
// It's because file's Content too large to pass it by value.
type File struct {
	Name    string
	Content []byte
}

// Extract extension from filename. Format: .ext
func GetExt(fname string) string {
	for i := len(fname) - 1; i >= 0; i-- {
		if fname[i] == '.' {
			return fname[i:]
		}
	}
	return ""
}

// Extract extension from File instance.
func (f *File) Extension() string {
	return GetExt(f.Name)
}

// Generate random filename, save original file's extension.
func (f *File) RandName() string {
	rand.Seed(time.Now().UnixNano())
	var (
		letters  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321"
		size     = rand.Int()%20 + 20
		randname = make([]byte, size)
	)
	for i := 0; i < size; i++ {
		randname[i] = letters[rand.Intn(len(letters))]
	}
	return string(randname) + GetExt(f.Name)
}

// Apply color mask on image, return new modified image.
func (f *File) Colorize() []byte {
	var (
		err  error
		cont []byte
	)
	switch f.Extension() {
	case ".png":
		cont, err = content.PngColorize(f.Content)
	case ".jpg":
		cont, err = content.JpegColorize(f.Content)
	default:
		break
	}
	if err != nil || cont == nil {
		return f.Content
	}
	return cont
}
