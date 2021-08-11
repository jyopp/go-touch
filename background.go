package main

import "image/color"

type Background struct {
	BasicLayer
	// Value from 0-255 controlling the brightness of the gradient
	Brightness int
}

func (background *Background) Init(frame Rect) *Background {
	background.BasicLayer.Init(frame, background)
	return background
}

func (background *Background) Draw(layer Layer, ctx DrawingContext) {
	bright := background.Brightness

	var c color.RGBA
	c.A = 0xFF

	//	g = byte((bright * 3) / 4)
	w, h := background.w, background.h
	row := make([]byte, 4*w)

	rect := background.Bounds()
	for y := 0; y < h; y++ {
		c.B = byte(bright * y / h)

		cornerL := rect.roundRectInset(y)
		cornerR := w - cornerL - 1
		for x := 0; x < w; x++ {
			i := 4 * x
			pix := row[i : i+4 : i+4]
			if x < cornerL || x > cornerR {
				pix[0], pix[1], pix[2], pix[3] = 0, 0, 0, 0xFF
			} else {
				c.R = byte(bright * x / w)
				c.G = byte(bright) - c.R/4 - c.B/2
				pix[0], pix[1], pix[2], pix[3] = c.R, c.G, c.B, c.A
			}
		}
		// Black out rounded corners on the background
		ctx.DrawRow(row, background.x, background.y+y)
	}
}
