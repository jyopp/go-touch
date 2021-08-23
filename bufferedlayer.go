package fbui

import (
	"image"
	"image/draw"
)

type BufferedLayer struct {
	BasicLayer
	Buffer
	invalid RegionList
}

func (layer *BufferedLayer) SetFrame(frame image.Rectangle) {
	if !layer.Eq(frame) {
		layer.BasicLayer.SetFrame(frame)
		layer.Buffer.SetFrame(frame)
	}
}

func (layer *BufferedLayer) InvalidateRect(rect image.Rectangle) {
	layer.invalid.AddRect(rect)
	layer.BasicLayer.InvalidateRect(rect)
}

func (layer *BufferedLayer) RenderBuffer() {
	buffer := &layer.Buffer
	for _, rect := range layer.invalid.Dequeue() {
		layer.BasicLayer.Render(buffer.Clip(rect))
	}
}

// Render draws and composites any invalid regions to the buffer
func (layer *BufferedLayer) Render(ctx DrawingContext) {
	layer.RenderBuffer()

	if rect := layer.Buffer.Rect.Intersect(ctx.Bounds()); !rect.Empty() {
		draw.Draw(ctx.Image(), rect, layer.Buffer.RGBA, rect.Min, draw.Over)
		ctx.SetDirty(rect)
	}
}
