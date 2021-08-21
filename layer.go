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

type LayerDrawDelegate interface {
	Draw(ctx DrawingContext)
}

type BasicLayer struct {
	image.Rectangle
	Radius     int
	Background color.Color
	Self       Layer

	parent       Layer
	children     []Layer
	needsDisplay bool
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
	layer.needsDisplay = true
	layer.Rectangle = frame
}

func (layer *BasicLayer) AddChild(child Layer) {
	child.SetParent(layer.Layer())
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
	rect := child.Frame()
	for _, other := range layer.children[index:] {
		if other.Frame().Overlaps(rect) {
			child.SetNeedsDisplay()
		}
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
	layer.SetNeedsDisplay()
	layer.parent = parent
}

func (layer *BasicLayer) Children() []Layer {
	return layer.children
}

func (layer *BasicLayer) NeedsDisplay() bool {
	return layer.needsDisplay
}

func (layer *BasicLayer) SetNeedsDisplay() {
	layer.needsDisplay = true
	if parent := layer.parent; parent != nil {
		// TODO: Just send the global rect to the Display.
		// This should take the form NeedsRedraw(rect)
		// Or perhaps Display.NeedsRedraw(layer) / Display.WillMove(layer)
		if !layer.IsOpaque() {
			parent.SetNeedsDisplay()
		}
	}
}

// DrawChildren draws child layers IFF they are visible in ctx, and (need display or overlap rect)
func (layer *BasicLayer) DrawChildren(ctx DrawingContext, mustDraw image.Rectangle) {
	// Restrict mustDraw to be within ctx for performance; Does not affect correctness.
	mustDraw = ctx.Bounds().Intersect(mustDraw)

	for _, child := range layer.children {
		if child.NeedsDisplay() || mustDraw.Overlaps(child.Frame()) {
			if clipped := ctx.Clip(child.Frame()); clipped != nil {
				child.Display(clipped)
			}
		} else {
			// TEMP: Always delegate down the hierarchy, until we have global invalidated rects
			if clipped := ctx.Clip(child.Frame()); clipped != nil {
				child.DisplayIfNeeded(ctx)
			}
		}
	}
}

func (layer *BasicLayer) DisplayIfNeeded(ctx DrawingContext) {
	if layer.needsDisplay {
		// When calling interface methods, call from outermost
		// struct type so that embedding types can override methods.
		layer.Layer().Display(ctx)
	} else {
		// Pass empty rectangle so no child is forcibly redrawn;
		// Only draws children with NeedsDisplay == true
		layer.DrawChildren(ctx, image.Rectangle{})
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
	// fmt.Printf("Drawing %T into %T %v\n", layer.Self, ctx, ctx.Bounds())

	// Eventually we'll need to convert into the destination coordinate space
	if drawer, ok := layer.Layer().(LayerDrawDelegate); ok {
		drawer.Draw(ctx)
	}

	// TODO: Let delegates decide what to mark dirty
	ctx.SetDirty(layer.Rectangle)
	// Redraw all visible children
	layer.DrawChildren(ctx, layer.Rectangle)
	layer.needsDisplay = false
}
