package main

import "image/color"

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

func pixel565(r, g, b byte) (byte, byte) {
	return ((g << 3) & 0b11100000) | b>>3, (r & 0b11111000) | (g >> 5)
}

func (model *ColorModel565) RGB(r, g, b byte) Color565 {
	b1, b2 := pixel565(r, g, b)
	return Color565{b1, b2}
}

func (model *ColorModel565) Convert(c color.Color) color.Color {
	r, g, b, _ := c.RGBA()
	var converted Color565
	converted.b1, converted.b2 = pixel565(byte(r), byte(g), byte(b))
	return converted
}

func (model *ColorModel565) FillRGB(ctx LayerDrawing, rect Rect, r, g, b byte) {
	b1, b2 := pixel565(r, g, b)
	w2 := rect.w * 2
	row := make([]byte, w2)

	for i := 0; i < w2; i += 2 {
		row[i], row[i+1] = b1, b2
	}

	// Copy the pixel row into all relevant output lines
	if rect.radius > 0 && !ctx.IsBuffered() {
		// Clip the corners when drawing into an unbuffered context
		for y := 0; y < rect.h; y++ {
			i := rect.roundRectInset(y)
			ctx.DrawRow(row[2*i:w2-2*i], rect.x+i, rect.y+y)
		}
	} else {
		for y := 0; y < rect.h; y++ {
			ctx.DrawRow(row, rect.x, rect.y+y)
		}
	}
}
