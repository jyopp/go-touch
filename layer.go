package main

import "image/color"

type Layer interface {
	Children() []Layer
	AddChild(Layer)
	InsertChild(Layer, int)

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
	layer.needsDisplay = layer.needsDisplay || layer.Rect != frame
	layer.Rect = frame
}

func (layer *BasicLayer) AddChild(child Layer) {
	layer.children = append(layer.children, child)
}

func (layer *BasicLayer) InsertChild(child Layer, index int) {
	if index < len(layer.children) {
		// Splice array into itself, duplicating element at index.
		layer.children = append(layer.children[:index+1], layer.children[index:]...)
		layer.children[index] = child
	} else {
		layer.children = append(layer.children, child)
	}
	for _, child := range layer.children[index:] {
		child.SetNeedsDisplay()
	}
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
		limeGreen := color.RGBA{R: 0, G: 0xFF, B: 0, A: 0xFF}
		ctx.Fill(rect, limeGreen)
	}

	// TODO: Let delegates decide what to mark dirty
	ctx.SetDirty(layer.Rect)

	for _, child := range layer.children {
		child.Display(ctx.Clip(child.Frame()))
	}

	layer.needsDisplay = false
}
