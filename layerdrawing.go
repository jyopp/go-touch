package main

import (
	"image"
	"image/color"
)

type Color565 struct {
	b1, b2 byte
}

func (c Color565) RGBA() (r, g, b, a uint32) {
	r = uint32(c.b1 & 0b11111000)
	g = uint32(c.b1<<5 | (c.b2>>5)<<2)
	b = uint32(c.b2 << 3)
	a = 0xFFFF
	return
}

type ColorModel565 struct{}

var model565 ColorModel565

func (model *ColorModel565) Convert(c color.Color) color.Color {
	r, g, b, _ := c.RGBA()
	var converted Color565
	converted.b1, converted.b2 = pixel565(byte(r), byte(g), byte(b))
	return converted
}

// This file contains methods to make layer an image.Image for image and font drawing

func (layer *Layer) Set(x, y int, c color.Color) {
	offset := 2 * (y*int(layer.w) + x)
	if c565, ok := c.(*Color565); ok {
		layer.Contents[offset], layer.Contents[offset+1] = c565.b1, c565.b2
	} else {
		r, g, b, _ := c.RGBA()
		layer.Contents[offset], layer.Contents[offset+1] = pixel565(byte(r), byte(g), byte(b))
	}
	layer.NeedsDisplay = true
}

func (layer *Layer) ColorModel() color.Model {
	return &model565
}

func (layer *Layer) Bounds() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{int(layer.w), int(layer.h)},
	}
}

func (layer *Layer) At(x, y int) color.Color {
	offset := 2 * (y*int(layer.w) + x)
	return Color565{layer.Contents[offset], layer.Contents[offset+1]}
}
