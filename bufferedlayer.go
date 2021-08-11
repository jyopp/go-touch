package main

import "image/draw"

type BufferedLayer struct {
	BasicLayer
	buffer      DisplayBuffer
	needsRedraw bool
}

// TODO: Return an error?
func (layer *BufferedLayer) Init(frame Rect, identity interface{}) {
	layer.BasicLayer.Init(frame, identity)
	layer.buffer = NewDisplayBuffer(nil, frame)
	layer.needsRedraw = true
}

func (layer *BufferedLayer) SetFrame(frame Rect) {
	layer.BasicLayer.SetFrame(frame)
	if layer.buffer.SetFrame(frame) {
		layer.needsRedraw = true
	}
}

func (layer *BufferedLayer) Display(ctx DrawingContext) {
	buffer := layer.buffer
	if layer.needsRedraw {
		layer.BasicLayer.Display(buffer)
	}

	min, max := buffer.Rect.Min, buffer.Rect.Max
	for y := min.Y; y < max.Y; y++ {
		row := buffer.GetRow(y)
		// Clip rounded corners in a very simple way
		if layer.radius > 0 {
			i := layer.roundRectInset(y - min.Y)
			ctx.DrawRow(row[4*i:len(row)-4*i], min.X+i, y, draw.Over)
		} else {
			ctx.DrawRow(row, min.X, y, draw.Over)
		}
	}
}
