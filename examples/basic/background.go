package main

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/jyopp/go-touch"
)

type Background struct {
	touch.BasicLayer
	// Value from 0-255 controlling the brightness of the gradient
	Brightness int
}

func (b *Background) Init(frame image.Rectangle, brightness int) {
	b.SetFrame(frame)
	b.Brightness = brightness
	b.Self = b
}

func (b *Background) DrawIn(ctx touch.DrawingContext) {
	bounds := ctx.Bounds()
	bright := b.Brightness

	var c color.RGBA
	c.A = 0xFF

	row := make([]byte, 4*bounds.Dx())
	origin, size := b.Rectangle.Min, b.Rectangle.Size()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		c.B = byte(bright * (y - origin.Y) / size.Y)

		// For each pixel in row
		for i := 0; i < len(row); i += 4 {
			c.R = byte(bright * (bounds.Min.X + i/4) / size.X)
			c.G = byte(bright) - c.R/4 - c.B/2

			pix := row[i : i+4 : i+4]
			pix[0], pix[1], pix[2], pix[3] = c.R, c.G, c.B, c.A
		}
		ctx.DrawRow(row, bounds.Min.X, y, draw.Src)
	}
	ctx.SetDirty(bounds)
}
