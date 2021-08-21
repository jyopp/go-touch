package fbui

import (
	"image"
	"image/draw"
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

	// BufferedLayer can draw its internal buffer without copying to a DrawingContext.
	if ctx != nil {
		// fmt.Printf("Compositing %T %v into %T %v\n", layer.Delegate, buffer.Rect, ctx, ctx.Bounds())
		draw.Draw(ctx.Image(), buffer.Rect, buffer, buffer.Rect.Min, draw.Over)
		ctx.SetDirty(buffer.Rect)
	}
}

func (layer *BufferedLayer) DisplayIfNeeded(ctx DrawingContext) {
	if layer.needsDisplay {
		// When calling interface methods, call from outermost
		// struct type so that embedding types can override methods.
		layer.Layer().Display(ctx)
	} else {
		for _, child := range layer.children {
			// Draw children into the buffer
			if clip := layer.Buffer.Clip(child.Frame()); clip != nil {
				child.DisplayIfNeeded(clip)
			}
		}
	}
}
