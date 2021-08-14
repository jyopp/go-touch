package main

import (
	"image"
	"image/color"
	"image/draw"
)

type Background struct {
	BasicLayer
	// Value from 0-255 controlling the brightness of the gradient
	Brightness int
}

func (background *Background) Init(frame image.Rectangle) *Background {
	background.BasicLayer.Init(frame, background)
	return background
}

func (background *Background) Draw(layer Layer, ctx DrawingContext) {
	rect := background.Rectangle
	bright := background.Brightness

	var c color.RGBA
	c.A = 0xFF

	h := rect.Dy()
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		c.B = byte(bright * (y - rect.Min.Y) / h)

		row := ctx.GetRow(rect.Min.Y + y)
		// For each pixel in row
		for i, rowLen := 0, len(row); i < rowLen; i += 4 {
			c.R = byte(bright * i / rowLen)
			c.G = byte(bright) - c.R/4 - c.B/2

			pix := row[i : i+4 : i+4]
			pix[0], pix[1], pix[2], pix[3] = c.R, c.G, c.B, c.A
		}
		ctx.DrawRow(row, rect.Min.X, y, draw.Src)
	}
	// Mask corners out with opaque black
	CornerMask{rect, background.Radius}.EraseCorners(ctx.Image())
}
