package main

type BufferedLayer struct {
	BasicLayer
	buffer      *LayerImageBuffer
	needsRedraw bool
}

// TODO: Return an error?
func (layer *BufferedLayer) Init(frame Rect, identity interface{}) {
	layer.BasicLayer.Init(frame, identity)
	layer.buffer = NewLayerImageBuffer(frame.w, frame.h)
	layer.needsRedraw = true
}

func (layer *BufferedLayer) SetFrame(frame Rect) {
	if layer.buffer != nil && (frame.w != layer.buffer.Width || frame.h != layer.buffer.Height) {
		layer.buffer = NewLayerImageBuffer(frame.w, frame.h)
		layer.needsDisplay = true
		layer.needsRedraw = true
	}
	layer.BasicLayer.SetFrame(frame)
}

func (layer *BufferedLayer) Display(ctx DrawingContext) {
	buffer := layer.buffer
	if layer.needsRedraw {
		layer.BasicLayer.Display(buffer)
	}

	for contentY := 0; contentY < buffer.Height; contentY++ {
		row := buffer.GetRow(contentY)
		// Clip rounded corners in a very simple way
		if layer.radius > 0 {
			i := layer.roundRectInset(contentY)
			ctx.DrawRow(row[2*i:len(row)-2*i], i, contentY)
		} else {
			ctx.DrawRow(buffer.GetRow(contentY), 0, contentY)
		}
	}
}
