package main

import (
	"bytes"
	"embed"
	"image"

	"image/png"

	"github.com/jyopp/fbui"
)

//go:embed images fonts
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

// RegisterFont registers a truetype font for on-demand loding
func (r *resourceReader) RegisterFont(name string) {
	fbui.RegisterTTF(name, func() []byte {
		data, _ := _resourceFiles.ReadFile("fonts/" + name)
		return data
	})
}
