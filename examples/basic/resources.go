package main

import (
	"bytes"
	"embed"
	"image"

	"image/draw"
	"image/png"

	"github.com/jyopp/go-touch"
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

func (r *resourceReader) ReadPNGTemplate(name string) (*image.Alpha, error) {
	orig, err := r.ReadPNG(name)
	if err != nil {
		return nil, err
	}
	alpha := image.NewAlpha(orig.Bounds())
	draw.Draw(alpha, alpha.Rect, orig, orig.Bounds().Min, draw.Src)
	return alpha, nil
}

// RegisterFont registers a truetype font for on-demand loading
func (r *resourceReader) RegisterFont(name string) {
	touch.RegisterTTF(name, func() []byte {
		data, _ := _resourceFiles.ReadFile("fonts/" + name)
		return data
	})
}
