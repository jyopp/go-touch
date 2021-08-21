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
	if layer.Eq(frame) {
		return
	}
	layer.BasicLayer.SetFrame(frame)
	layer.Buffer.SetFrame(frame)
}

func (layer *BufferedLayer) InvalidateRect(rect image.Rectangle) {
	layer.invalid.AddRect(rect)
	layer.BasicLayer.InvalidateRect(rect)
}

// Render draws and composites any invalid regions to the buffer
func (layer *BufferedLayer) Render() {
	for _, rect := range layer.invalid.Dequeue() {
		layer.Layer().DrawIn(layer.Buffer.Clip(rect))
	}
}

// DrawIn redraws the layer and its children into the buffer.
// Unlike other layers, ctx may be nil; In this case, the buffer is drawn
// but not copied to a drawing context.
func (layer *BufferedLayer) DrawIn(ctx DrawingContext) {
	// Delegate when drawing into our buffer
	if layer.Buffer.IsAncestor(ctx) {
		layer.BasicLayer.DrawIn(ctx)
		return
	}

	// When rendering to an external context, update the buffer and draw from it
	layer.Render()

	if ctx != nil {
		rect := layer.Buffer.Bounds().Intersect(ctx.Bounds())
		draw.Draw(ctx.Image(), rect, layer.Buffer.RGBA, rect.Min, draw.Over)
		ctx.SetDirty(rect)
	}
}
