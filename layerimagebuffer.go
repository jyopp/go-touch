package main

import (
	"image"
	"image/color"
)

type LayerImageBuffer struct {
	Width, Height int
	// Pixel array of 16-bit pixels in big-endian rgb 565 order
	pixels []byte
	// Allow Buffers to be shared with different clipping rects
	rect Rect
}

func NewLayerImageBuffer(w, h int) *LayerImageBuffer {
	return &LayerImageBuffer{
		Width:  w,
		Height: h,
		pixels: make([]byte, 2*w*h),
		rect:   Rect{0, 0, w, h, 0},
	}
}

func (layer *LayerImageBuffer) Clip(rect Rect) DrawingContext {
	clone := *layer
	clone.rect = clone.rect.Intersection(rect)
	return &clone
}

// LayerImageBuffer conforms to LayerDrawing

func (layer *LayerImageBuffer) Set(x, y int, c color.Color) {
	var ok bool
	var c565 Color565
	if c565, ok = c.(Color565); !ok {
		c565 = model565.Convert(c).(Color565)
	}
	x += layer.rect.x
	y += layer.rect.y
	offset := 2 * (y*layer.Width + x)
	layer.pixels[offset], layer.pixels[offset+1] = c565.b1, c565.b2
}

func (layer *LayerImageBuffer) ColorModel() color.Model {
	return model565
}

func (layer *LayerImageBuffer) Bounds() image.Rectangle {
	return layer.rect.Rectangle()
}

func (layer *LayerImageBuffer) At(x, y int) color.Color {
	x += layer.rect.x
	y += layer.rect.y
	offset := 2 * (y*layer.Width + x)
	return Color565{layer.pixels[offset], layer.pixels[offset+1]}
}

func (layer *LayerImageBuffer) IsBuffered() bool {
	return true
}

func (layer *LayerImageBuffer) GetRow(y int) []byte {
	y += layer.rect.y
	rowOffset := 2 * (layer.Width*y + layer.rect.x)
	return layer.pixels[rowOffset : rowOffset+2*layer.rect.w]
}

func (layer *LayerImageBuffer) DrawRow(row []byte, x, y int) {
	// Bounds-check before doing any real work
	if y < 0 || y >= layer.rect.Bottom() {
		return
	}
	if x < 0 {
		row = row[-x:]
		x = -x
	}

	x += layer.rect.x
	y += layer.rect.y

	if len(row) >= 2*layer.rect.w {
		row = row[:len(row)-2*x]
	}

	dst := layer.pixels[2*(layer.Width*y+x):]
	copy(dst, row)
}
