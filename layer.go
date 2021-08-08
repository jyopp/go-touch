package main

type Layer interface {
	Children() []Layer
	AddChild(Layer)

	Frame() Rect
	SetFrame(frame Rect)

	NeedsDisplay() bool
	SetNeedsDisplay()
	DisplayIfNeeded(ctx DrawingContext)
	Display(ctx DrawingContext)

	HitTest(TouchEvent) TouchTarget
}

type TouchTarget interface {
	StartTouch(TouchEvent)
	UpdateTouch(TouchEvent)
	EndTouch(TouchEvent)
}

type LayerDrawer interface {
	Draw(layer Layer, ctx DrawingContext)
}

type BasicLayer struct {
	Rect
	children     []Layer
	needsDisplay bool
	identity     interface{}
}

func (layer *BasicLayer) Init(frame Rect, identity interface{}) {
	layer.Rect = frame
	layer.identity = identity
	layer.needsDisplay = true
}

func (layer *BasicLayer) Frame() Rect {
	return layer.Rect
}

func (layer *BasicLayer) SetFrame(frame Rect) {
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

func (layer *BasicLayer) Children() []Layer {
	return layer.children
}

func (layer *BasicLayer) NeedsDisplay() bool {
	return layer.needsDisplay
}

func (layer *BasicLayer) SetNeedsDisplay() {
	layer.needsDisplay = true
}

func (layer *BasicLayer) DisplayIfNeeded(ctx DrawingContext) {
	if layer.needsDisplay {
		layer.Display(ctx)
	} else {
		for _, child := range layer.children {
			child.DisplayIfNeeded(ctx.Clip(child.Frame()))
		}
	}
}

// Display naively splats all of the buffer's pixels into the parent's content
func (layer *BasicLayer) Display(ctx DrawingContext) {
	// Eventually we'll need to convert into the destination coordinate space
	// fmt.Printf("Drawing %T %v into %T %v\n", layer, layer, ctx, ctx)
	if drawer, ok := layer.identity.(LayerDrawer); ok {
		drawer.Draw(layer, ctx)
	} else {
		// TODO: Throw an error? Refuse to draw?
		// Draw lime green for debugging
		rect := Rect{x: 0, y: 0, w: layer.w, h: layer.h}
		model565.FillRGB(ctx, rect, 0x00, 0xFF, 0x00)
	}

	for _, child := range layer.children {
		child.Display(ctx.Clip(child.Frame()))
	}
	layer.needsDisplay = false
}
