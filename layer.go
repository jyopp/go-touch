package fbui

import (
	"image"
	"image/color"
)

type Layer interface {
	Children() []Layer
	AddChild(Layer)
	InsertChild(Layer, int)

	Frame() image.Rectangle
	SetFrame(frame image.Rectangle)

	Invalidate()
	InvalidateRect(rect image.Rectangle)
	DrawIn(ctx DrawingContext)

	HitTest(TouchEvent) LayerTouchDelegate

	Parent() Layer
	SetParent(Layer)

	// TODO/WIP: Add IsOpaque()
	IsOpaque() bool
	// In a pre-rendering phase, collect the UNION of all child
	// frames that need to be displayed with DrawingMode draw.Copy;
	// In the rendering phase, draw dirty views in the global context,
	// and (if the union rect is nonempty) draw all nondirty views
	// that overlap the union rect in a clipped context.
	// We may also choose to build a drawlist in which views that
	// would be completely covered by another, opaque view are filtered
	// out.
}

type LayerTouchDelegate interface {
	StartTouch(TouchEvent)
	UpdateTouch(TouchEvent)
	EndTouch(TouchEvent)
}

type BasicLayer struct {
	image.Rectangle
	Radius     int
	Background color.Color
	Self       Layer

	parent   Layer
	children []Layer
}

// Layer returns a layer interface to the outermost struct associated with this layer.
func (layer *BasicLayer) Layer() Layer {
	if layer.Self != nil {
		return layer.Self
	}
	return layer
}

func (layer *BasicLayer) Frame() image.Rectangle {
	return layer.Rectangle
}

func (layer *BasicLayer) SetFrame(frame image.Rectangle) {
	if layer.Eq(frame) {
		return
	}
	leafClass := layer.Layer()
	leafClass.InvalidateRect(layer.Rectangle)
	leafClass.InvalidateRect(frame)

	layer.Rectangle = frame
}

func (layer *BasicLayer) AddChild(child Layer) {
	child.SetParent(layer.Layer())
	layer.children = append(layer.children, child)
}

func (layer *BasicLayer) InsertChild(child Layer, index int) {
	child.SetParent(layer.Layer())
	if index < len(layer.children) {
		// Splice array into itself, duplicating element at index.
		layer.children = append(layer.children[:index+1], layer.children[index:]...)
		layer.children[index] = child
	} else {
		layer.children = append(layer.children, child)
	}
}

func (layer *BasicLayer) HitTest(event TouchEvent) LayerTouchDelegate {
	for _, child := range layer.children {
		if target := child.HitTest(event); target != nil {
			return target
		}
	}
	if event.Pressed && event.In(layer.Rectangle) {
		if interactor, ok := layer.Self.(LayerTouchDelegate); ok {
			return interactor
		}
	}
	return nil
}

func (layer *BasicLayer) IsOpaque() bool {
	if layer.Background != nil {
		_, _, _, a := layer.Background.RGBA()
		return a == 0xFFFF
	}
	return false
}

func (layer *BasicLayer) Parent() Layer {
	return layer.parent
}

func (layer *BasicLayer) SetParent(parent Layer) {
	layer.parent = parent
	layer.Invalidate()
}

func (layer *BasicLayer) Children() []Layer {
	return layer.children
}

func (layer *BasicLayer) Invalidate() {
	layer.Layer().InvalidateRect(layer.Rectangle)
}

func (layer *BasicLayer) InvalidateRect(rect image.Rectangle) {
	if p := layer.Parent(); p != nil {
		p.InvalidateRect(rect)
	}
}

// DrawChildren draws child layers IFF they are visible in ctx, and (need display or overlap rect)
func (layer *BasicLayer) DrawChildren(ctx DrawingContext) {
	// Restrict mustDraw to be within ctx for performance; Does not affect correctness.
	drawRect := ctx.Bounds()
	for _, child := range layer.children {
		if drawRect.Overlaps(child.Frame()) {
			if clipped := ctx.Clip(child.Frame()); clipped != nil {
				child.DrawIn(clipped)
			}
		}
	}
}

// For delegation of default drawing behavior (Background / roundrect)
func (layer *BasicLayer) DrawIn(ctx DrawingContext) {
	if layer.Background != nil {
		ctx.Fill(layer.Rectangle, layer.Background, layer.Radius)
	}
	layer.DrawChildren(ctx)
}
