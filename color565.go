package main

import "image/color"

type Color565 struct {
	b1, b2 byte
}

func (c Color565) RGBA() (r, g, b, a uint32) {
	r = uint32(c.b2&0b11111000) << 8
	g = uint32(c.b2<<5|(c.b1>>5)<<2) << 8
	b = uint32(c.b1<<3) << 8
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
	var converted Color565
	r, g, b, a := c.RGBA()
	if a > 0 {
		// Unpremultiply alpha
		r = 0xFFFF * r / a
		g = 0xFFFF * g / a
		b = 0xFFFF * b / a
	}
	converted.b1, converted.b2 = pixel565(byte(r>>8), byte(g>>8), byte(b>>8))
	return converted
}

func (model *ColorModel565) Fill(ctx DrawingContext, rect Rect, color color.Color) {
	data := model.Convert(color).(Color565)
	w2 := rect.w * 2
	row := make([]byte, w2)

	for i := 0; i < w2; i += 2 {
		row[i], row[i+1] = data.b1, data.b2
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
