package fbui

import (
	"image"
	"image/color"
)

type Layer interface {
	Children() []Layer
	AddChild(...Layer)
	InsertChild(Layer, int)
	RemoveChild(Layer)

	Frame() image.Rectangle
	SetFrame(frame image.Rectangle)

	Invalidate()
	InvalidateRect(rect image.Rectangle)

	// Render should render this layer and its subtree into ctx
	Render(ctx DrawingContext)

	// DrawIn should draw the parts of this layer that are visible in ctx
	DrawIn(ctx DrawingContext)

	HitTest(TouchEvent) LayerTouchDelegate

	Parent() Layer
	SetParent(Layer)
	RemoveFromParent()

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

type touchBlocker struct{}

func (tb touchBlocker) StartTouch(TouchEvent)  {}
func (tb touchBlocker) UpdateTouch(TouchEvent) {}
func (tb touchBlocker) EndTouch(TouchEvent)    {}

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
	if !layer.Rectangle.Empty() {
		leafClass.InvalidateRect(layer.Rectangle)
	}
	leafClass.InvalidateRect(frame)

	layer.Rectangle = frame
}

func (layer *BasicLayer) AddChild(layers ...Layer) {
	layer.children = append(layer.children, layers...)
	self := layer.Layer()
	for _, child := range layers {
		child.SetParent(self)
	}
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

func (layer *BasicLayer) RemoveChild(child Layer) {
	for idx := range layer.children {
		if layer.children[idx] == child {
			layer.children = append(layer.children[:idx], layer.children[idx+1:]...)
			layer.Layer().InvalidateRect(child.Frame())
			return
		}
	}
}

func (layer *BasicLayer) HitTest(event TouchEvent) LayerTouchDelegate {
	for idx := len(layer.children); idx > 0; idx-- {
		if target := layer.children[idx-1].HitTest(event); target != nil {
			return target
		}
	}
	if event.Pressed && event.In(layer.Rectangle) {
		if interactor, ok := layer.Self.(LayerTouchDelegate); ok {
			return interactor
		}
		// Visible views block touches by default
		// TODO: More-explicit method of blocking or allowing touches
		if layer.Background != nil {
			return touchBlocker{}
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

func (layer *BasicLayer) RemoveFromParent() {
	if layer.parent != nil {
		layer.parent.RemoveChild(layer.Layer())
	}
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
func (layer *BasicLayer) Render(ctx DrawingContext) {
	// Draw the smallest rect of this layer that is not occluded by opaque children
	rect := ctx.Bounds()
	for _, child := range layer.children {
		if child.IsOpaque() {
			rect = Winnow(rect, child.Frame())
		}
	}
	// Draw this layer IFF there are pixels to be drawn.
	if clipped := ctx.Clip(rect); !clipped.Bounds().Empty() {
		layer.Layer().DrawIn(clipped)
	}

	// Render children (separate phase)
	// TODO: Clip overlapping children of lower Z-order
	for _, child := range layer.children {
		if clipped := ctx.Clip(child.Frame()); !clipped.Bounds().Empty() {
			child.Render(clipped)
		}
	}
}

// For delegation of default drawing behavior (Background / roundrect)
func (layer *BasicLayer) DrawIn(ctx DrawingContext) {
	if layer.Background != nil {
		ctx.Fill(layer.Rectangle, layer.Background, layer.Radius)
	}
}
