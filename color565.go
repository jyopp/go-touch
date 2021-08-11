package main

import (
	"image/color"
)

type Color565 struct {
	b1, b2 byte
}

func (c Color565) RGBA() (r, g, b, a uint32) {
	// No precision is lost as long as the high bits are accurate.
	// Bits are duplicated to maximize coverage of the available gamut.

	// Duplicate 5 bits three times; Fill 15/16 bits, low bit always unset.
	r = (uint32(c.b2) >> 3) * 0b0000100001000010
	b = (uint32(c.b1) & 0b11111) * 0b0000100001000010
	// Stretching 6 green bits striped across two bytes is much harder...
	// Doubles 6 bits to fill into 12/16 bits, leaving 4 unset.
	// To use fewer ops, multiplication constant is downshifted 2 bits.
	g = (uint32(c.b2<<5|c.b1>>3) & 0b11111100) * 0b0000000100000100
	a = 0xFFFF
	return
}

type ColorModel565 struct{}

var model565 ColorModel565

func pixel565(r, g, b byte) (byte, byte) {
	return ((g << 3) & 0b11100000) | b>>3, (r & 0b11111000) | (g >> 5)
}

func (model ColorModel565) RGB(r, g, b byte) Color565 {
	b1, b2 := pixel565(r, g, b)
	return Color565{b1, b2}
}

var pixelPathCounts = []int{0, 0, 0}

// var pixelTypes = make(map[string]int)

func (model ColorModel565) DumpStats() {
	// fmt.Printf("ConversionPaths %v; Generic %v\n", pixelPathCounts, pixelTypes)
}

func (model ColorModel565) Convert(c color.Color) color.Color {
	// output color is implicitly initialized to black (0,0)
	var converted Color565

	switch cast := c.(type) {
	// Image drawing routines pass a pointer-to-color
	case *color.RGBA64:
		pixelPathCounts[2]++
		converted.b1, converted.b2 = pixel565(byte(cast.R>>8), byte(cast.G>>8), byte(cast.B>>8))
		return converted

	case color.NRGBA:
		if cast.A == 0xFF {
			// Ignore alpha for opaque colors.
			pixelPathCounts[0]++
			converted.b1, converted.b2 = pixel565(cast.R, cast.G, cast.B)
		} else if cast.A > 0 {
			// For partially-transparent colors expand precision and multiply alpha
			pixelPathCounts[1]++
			a := uint(cast.A)
			r, g, b := a*uint(cast.R)>>8, a*uint(cast.G)>>8, a*uint(cast.B)>>8
			converted.b1, converted.b2 = pixel565(byte(r), byte(g), byte(b))
		}
		return converted

	case color.RGBA:
		// Base case; RGB is premultiplied so just pack the bytes
		converted.b1, converted.b2 = pixel565(cast.R, cast.G, cast.B)
		return converted

	case Color565:
		// Other base case; Conversion was called by mistake?
		return cast

	default:
		// Trace unhandled types (optional)
		// pixelTypes[reflect.TypeOf(c).Name()]++

		// Fallback path; Read premultiplied values via interface
		r, g, b, _ := c.RGBA()

		converted.b1, converted.b2 = pixel565(byte(r>>8), byte(g>>8), byte(b>>8))
		return converted
	}
}

func (model ColorModel565) Fill(ctx DrawingContext, rect Rect, color color.Color) {
	data := model.Convert(color).(Color565)
	w2 := rect.w * 2
	row := make([]byte, w2)

	for i := 0; i < w2; i += 2 {
		row[i], row[i+1] = data.b1, data.b2
	}

	// Copy the pixel row into all relevant output lines
	if rect.radius > 0 {
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
