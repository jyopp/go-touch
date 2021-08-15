package fbui

import (
	"image"
	"image/color"
	"image/draw"
)

type ControlState int

const (
	StateNormal      ControlState = 0x0
	StateHighlighted ControlState = 0x1
	StateDisabled    ControlState = 0x80
)

type Button struct {
	BasicLayer
	OnTap func()
	Label TextLayer
	Icon  image.Image
	State ControlState
}

func (b *Button) Init(frame image.Rectangle, labelFont string, size float64) {
	b.SetFrame(frame)
	b.Background = color.Transparent
	b.Radius = 5
	b.Label.Init(image.Rectangle{}, labelFont, size)
	b.Label.Gravity = GravityCenter
	b.Delegate = b
	b.applyColors()
}

func (b *Button) SetState(state ControlState) {
	if state != b.State {
		b.State = state
		b.applyColors()
	}
}

func (b *Button) SetDisabled(disabled bool) {
	if disabled {
		b.SetState(StateDisabled)
	} else {
		b.SetState(StateNormal)
	}
}

func (b *Button) SetHighlighted(highlighted bool) {
	if b.State == StateDisabled {
		return
	} else if highlighted {
		b.SetState(StateHighlighted)
	} else {
		b.SetState(StateNormal)
	}
}

func (b *Button) applyColors() {
	switch b.State {
	case StateNormal:
		b.Background = color.RGBA{R: 0xFF, G: 0xFE, B: 0xFC, A: 0xFF}
		b.Label.Color = color.Black
	case StateHighlighted:
		b.Background = color.RGBA{R: 0x66, G: 0x99, B: 0xCC, A: 0xFF}
		b.Label.Color = color.White
	case StateDisabled:
		b.Background = color.RGBA{R: 0xBB, G: 0xBB, B: 0xBB, A: 0xDD}
		b.Label.Color = color.Gray{0x77}
	}
	b.needsDisplay = true
}

// Need some sort of prepare phase for drawing
func (b *Button) Draw(layer Layer, ctx DrawingContext) {
	ctx.Fill(b.Rectangle, b.Background, b.Radius, draw.Over)

	layout := LayoutRect{b.Rectangle.Inset(8)}

	if b.Icon != nil {
		iconSize := b.Icon.Bounds().Size()
		iconBounds := layout.Slice(iconSize.Y, 5, FromTop).Aligned(iconSize, GravityCenter)
		draw.Draw(ctx.Image(), iconBounds, b.Icon, image.Point{}, draw.Over)
	}

	// Render the label text in the remaining area
	b.Label.SetFrame(layout.Rectangle)
	b.Label.Draw(&b.Label, ctx)
}

func (b *Button) StartTouch(event TouchEvent) {
	b.SetHighlighted(event.In(b.Rectangle))
}

func (b *Button) UpdateTouch(event TouchEvent) {
	b.SetHighlighted(event.In(b.Rectangle))
}

func (b *Button) EndTouch(event TouchEvent) {
	if event.In(b.Rectangle) && b.OnTap != nil {
		b.OnTap()
	}
	b.SetHighlighted(false)
}
