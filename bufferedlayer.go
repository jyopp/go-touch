package main

import (
	"image"
	"image/draw"
)

type BufferedLayer struct {
	BasicLayer
	buffer      *DisplayBuffer
	needsRedraw bool
}

func (layer *BufferedLayer) SetFrame(frame image.Rectangle) {
	if layer.Eq(frame) {
		return
	}
	layer.BasicLayer.SetFrame(frame)
	if layer.buffer == nil {
		layer.buffer = NewDisplayBuffer(nil, frame)
		layer.needsRedraw = true
	} else if layer.buffer.SetFrame(frame) {
		layer.needsRedraw = true
	}
}

func (layer *BufferedLayer) Display(ctx DrawingContext) {
	buffer := layer.buffer
	if layer.needsRedraw {
		layer.BasicLayer.Display(buffer)
	}

	// fmt.Printf("Compositing %T %v into %T %v\n", layer.identity, buffer.Rect, ctx, ctx.Bounds())
	draw.Draw(ctx.Image(), buffer.Rect, buffer, buffer.Rect.Min, draw.Over)
	ctx.SetDirty(buffer.Rect)
}
