package main

import (
	"image"
	"image/color"
)

type LayerImageBuffer struct {
	Width, Height int
	// Pixel array of 16-bit pixels in big-endian rgb 565 order
	pixels []byte
}

func NewLayerImageBuffer(w, h int) *LayerImageBuffer {
	return &LayerImageBuffer{
		Width:  w,
		Height: h,
		pixels: make([]byte, 2*w*h),
	}
}

// LayerImageBuffer conforms to LayerDrawing

func (layer *LayerImageBuffer) Set(x, y int, c color.Color) {
	var ok bool
	var c565 Color565
	if c565, ok = c.(Color565); !ok {
		c565 = model565.Convert(c).(Color565)
	}
	offset := 2 * (y*layer.Width + x)
	layer.pixels[offset], layer.pixels[offset+1] = c565.b1, c565.b2
}

func (layer *LayerImageBuffer) ColorModel() color.Model {
	return &model565
}

func (layer *LayerImageBuffer) Bounds() (rect image.Rectangle) {
	rect.Max.X = layer.Width
	rect.Max.Y = layer.Height
	return
}

func (layer *LayerImageBuffer) At(x, y int) color.Color {
	offset := 2 * (y*layer.Width + x)
	return Color565{layer.pixels[offset], layer.pixels[offset+1]}
}

func (layer *LayerImageBuffer) IsBuffered() bool {
	return true
}

func (layer *LayerImageBuffer) GetRow(y int) []byte {
	lLen := 2 * layer.Width
	return layer.pixels[lLen*y : lLen*y+lLen]
}

func (layer *LayerImageBuffer) DrawRow(row []byte, x, y int) {
	// Bounds-check before doing any real work
	if y < 0 || y >= layer.Height {
		return
	}
	if x < 0 {
		row = row[-x:]
		x = -x
	}

	lineLen := 2 * layer.Width
	byteX := 2 * x
	if byteX+len(row) >= lineLen {
		row = row[:lineLen-byteX]
	}
	copy(layer.pixels[y*lineLen+byteX:], row)
}

func (layer *LayerImageBuffer) FillRGB(r, g, b byte) {
	b1, b2 := pixel565(r, g, b)
	w2 := layer.Width * 2
	for i := 0; i < w2; {
		layer.pixels[i] = b1
		i++
		layer.pixels[i] = b2
		i++
	}
	// Copy the first line into all subsequent lines
	firstLine := layer.pixels[:w2]
	for i := 1; i < layer.Height; i++ {
		copy(layer.pixels[i*w2:], firstLine)
	}
}
