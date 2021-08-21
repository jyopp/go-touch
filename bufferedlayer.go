package fbui

import (
	"image"
)

type BufferedLayer struct {
	BasicLayer
	Buffer
}

func (layer *BufferedLayer) SetFrame(frame image.Rectangle) {
	if layer.Eq(frame) {
		return
	}
	layer.BasicLayer.SetFrame(frame)
	layer.Buffer.SetFrame(frame)
}

// Display redraws the layer and its children into the buffer.
// Unlike other layers, ctx may be nil; In this case, the buffer is drawn
// but not copied to a drawing context.
func (layer *BufferedLayer) Display(ctx DrawingContext) {
	buffer := &layer.Buffer
	if layer.needsDisplay {
		layer.BasicLayer.Display(buffer)
	}

	if ctx != nil {
		buffer.SetDirty(ctx.Bounds())
		buffer.Flush(ctx)
	}
}

func (layer *BufferedLayer) DisplayIfNeeded(ctx DrawingContext) {
	if layer.needsDisplay {
		// When calling interface methods, call from outermost
		// struct type so that embedding types can override methods.
		layer.Layer().Display(ctx)
	} else {
		layer.DrawChildren(&layer.Buffer, image.Rectangle{})
		layer.Buffer.Flush(ctx)
	}
}
