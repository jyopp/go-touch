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

	NeedsDisplay() bool
	SetNeedsDisplay()
	DisplayIfNeeded(ctx DrawingContext)
	Display(ctx DrawingContext)

	HitTest(TouchEvent) LayerTouchDelegate

	// TODO: Add IsOpaque()
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

type LayerDrawDelegate interface {
	Draw(ctx DrawingContext)
}

type BasicLayer struct {
	image.Rectangle
	Radius     int
	Background color.Color
	Delegate   interface{}

	children     []Layer
	needsDisplay bool
}

func (layer *BasicLayer) Frame() image.Rectangle {
	return layer.Rectangle
}

func (layer *BasicLayer) SetFrame(frame image.Rectangle) {
	if layer.Eq(frame) {
		return
	}
	layer.needsDisplay = true
	layer.Rectangle = frame
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

func (layer *BasicLayer) HitTest(event TouchEvent) LayerTouchDelegate {
	for _, child := range layer.children {
		if target := child.HitTest(event); target != nil {
			return target
		}
	}
	if event.Pressed && event.In(layer.Rectangle) {
		if interactor, ok := layer.Delegate.(LayerTouchDelegate); ok {
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
		// When calling interface methods, call from outermost
		// struct type so that embedding types can override methods.
		if l, ok := layer.Delegate.(Layer); ok {
			l.Display(ctx)
		} else {
			layer.Display(ctx)
		}
	} else {
		for _, child := range layer.children {
			if clip := ctx.Clip(child.Frame()); clip != nil {
				child.DisplayIfNeeded(clip)
			}
		}
	}
}

// For delegation of default drawing behavior (Background / roundrect)
func (layer *BasicLayer) Draw(ctx DrawingContext) {
	if layer.Background != nil {
		ctx.Fill(layer.Rectangle, layer.Background, layer.Radius)
	}
}

// Display redraws the layer and its sublayers as needed, directly into ctx
func (layer *BasicLayer) Display(ctx DrawingContext) {
	// fmt.Printf("Drawing %T into %T %v\n", layer.identity, ctx, ctx.Bounds())

	layerRect := layer.Rectangle
	// Eventually we'll need to convert into the destination coordinate space
	// fmt.Printf("Drawing %T %v into %T %v\n", layer, layer, ctx, ctx)
	if drawer, ok := layer.Delegate.(LayerDrawDelegate); ok {
		drawer.Draw(ctx)
	} else {
		layer.Draw(ctx)
	}

	// TODO: Let delegates decide what to mark dirty
	ctx.SetDirty(layerRect)

	for _, child := range layer.children {
		if clip := ctx.Clip(child.Frame()); clip != nil {
			if child.NeedsDisplay() || clip.Bounds().Overlaps(layerRect) {
				child.Display(clip)
			}
		}
	}

	layer.needsDisplay = false
}
