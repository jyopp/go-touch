package main

import (
	"bytes"
	"embed"
	"image"

	"image/png"
)

//go:embed images
var _resourceFiles embed.FS

type resourceReader struct{}

var Resources resourceReader

func (r *resourceReader) ReadPNG(name string) (image.Image, error) {
	data, err := _resourceFiles.ReadFile("images/" + name)
	if err != nil {
		return nil, err
	}
	return png.Decode(bytes.NewReader(data))
}
