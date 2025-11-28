package imagex

import (
	"bytes"
	"image"
	"io"

	"github.com/r3dpixel/toolkit/bytex"
	"github.com/sunshineplan/imgconv"
)

// From reads the contents of the reader as a decoded image
func From(r io.Reader) (image.Image, error) {
	return imgconv.Decode(r)
}

// FromFile reads the contents of the file at the specified path as a decoded image
func FromFile(path string) (image.Image, error) {
	return imgconv.Open(path)
}

// FromBytes reads the contents of the byte array as a decoded image
func FromBytes(b []byte) (image.Image, error) {
	return imgconv.Decode(bytes.NewReader(b))
}

// To writes the image source to the specified writer
func To(imageSource image.Image, w io.Writer, format imgconv.Format) error {
	return imgconv.Write(w, imageSource, &imgconv.FormatOption{Format: format})
}

// ToFile writes the image source at the specified file path
func ToFile(imageSource image.Image, path string, format imgconv.Format) error {
	return imgconv.Save(path, imageSource, &imgconv.FormatOption{Format: format})
}

// ToBytes writes the image source to a byte array
func ToBytes(imageSource image.Image, format imgconv.Format) ([]byte, error) {
	// Calculate the size of the buffer
	bounds := imageSource.Bounds()
	size := bytex.Size(bounds.Dx()*bounds.Dy()*3) * bytex.B
	// Create the buffer
	buf := bytes.NewBuffer(make([]byte, 0, size))
	// Write the image to the buffer
	if err := To(imageSource, buf, format); err != nil {
		return nil, err
	}
	// Return the buffer bytes
	return buf.Bytes(), nil
}
