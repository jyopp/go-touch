package main

type Layer interface {
	// Content() LayerDrawing

	// Parent() Layer
	// SetParent(parent Layer)

	Children() []Layer
	AddChild(Layer)

	Frame() Rect
	SetFrame(frame Rect)

	NeedsDisplay() bool
	SetNeedsDisplay()
	DisplayIfNeeded(ctx LayerDrawing)
	Display(ctx LayerDrawing)

	HitTest(TouchEvent) TouchTarget
}

type TouchTarget interface {
	StartTouch(TouchEvent)
	UpdateTouch(TouchEvent)
	EndTouch(TouchEvent)
}

type LayerDrawer interface {
	Draw(layer Layer, ctx LayerDrawing)
}

type BasicLayer struct {
	Rect
	buffer *LayerImageBuffer
	// parent       Layer
	children     []Layer
	needsDisplay bool
	needsRedraw  bool
	identity     interface{}
}

func NewLayer(frame Rect, identity interface{}) *BasicLayer {
	return &BasicLayer{
		Rect: frame,
		// buffer:       NewLayerImageBuffer(frame.w, frame.h),
		needsDisplay: true,
		needsRedraw:  true,
		identity:     identity,
	}
}

func (layer *BasicLayer) Frame() Rect {
	return layer.Rect
}

func (layer *BasicLayer) SetFrame(frame Rect) {
	// if layer.buffer != nil && (frame.w != layer.buffer.Width || frame.h != layer.buffer.Height) {
	// 	layer.buffer = NewLayerImageBuffer(frame.w, frame.h)
	// 	layer.needsDisplay = true
	// 	layer.needsRedraw = true
	// }
	layer.Rect = frame
}

func (layer *BasicLayer) AddChild(child Layer) {
	layer.children = append(layer.children, child)
	// child.SetParent(layer)
	layer.needsDisplay = true
}

func (layer *BasicLayer) HitTest(event TouchEvent) TouchTarget {
	for _, child := range layer.children {
		if target := child.HitTest(event); target != nil {
			return target
		}
	}
	if event.Pressed && layer.Contains(event.X, event.Y) {
		if interactor, ok := layer.identity.(TouchTarget); ok {
			return interactor
		}
	}
	return nil
}

// func (layer *BasicLayer) Parent() Layer {
// 	return layer.parent
// }

// func (layer *BasicLayer) SetParent(parent Layer) {
// 	layer.parent = parent
// 	layer.needsDisplay = true
// }

func (layer *BasicLayer) Children() []Layer {
	return layer.children
}

// func (layer *BasicLayer) Content() LayerDrawing {
// 	return layer.buffer
// }

func (layer *BasicLayer) NeedsDisplay() bool {
	return layer.needsDisplay
}

func (layer *BasicLayer) SetNeedsDisplay() {
	layer.needsDisplay = true
}

func (layer *BasicLayer) DisplayIfNeeded(ctx LayerDrawing) {
	if layer.needsDisplay {
		layer.Display(ctx)
	} else {
		for _, child := range layer.children {
			child.DisplayIfNeeded(ctx)
		}
	}
}

// Display naively splats all of the buffer's pixels into the parent's content
func (layer *BasicLayer) Display(ctx LayerDrawing) {
	// Eventually we'll need to convert into the destination coordinate space
	// fmt.Printf("Drawing %T %v into %T %v\n", layer, layer, ctx, ctx)

	var x, y = layer.x, layer.y
	if drawer, ok := layer.identity.(LayerDrawer); ok {
		var into LayerDrawing
		if layer.buffer != nil {
			into = layer.buffer
		} else {
			into = ctx
		}
		drawer.Draw(layer, into)
	}

	if buffer := layer.buffer; buffer != nil {
		for contentY := 0; contentY < buffer.Height; contentY++ {
			row := buffer.GetRow(contentY)
			// Clip rounded corners in a very simple way
			if layer.radius > 0 {
				i := layer.roundRectInset(contentY)
				ctx.DrawRow(row[2*i:len(row)-2*i], x+i, y)
			} else {
				ctx.DrawRow(buffer.GetRow(contentY), x, y)
			}

			y++
		}
	}
	for _, child := range layer.children {
		child.Display(ctx)
	}
	layer.needsDisplay = false
}
