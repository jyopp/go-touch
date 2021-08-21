package fbui

import (
	"image"
	"image/color"
)

type Button struct {
	ControlLayer
	Label   TextLayer
	Icon    ImageLayer
	Actions [ControlActionsCount]func(*Button)

	Spacing int // Min. Distance between Icon and Label
}

func (b *Button) Init(frame image.Rectangle, labelFont string, size float64) {
	b.SetFrame(frame)
	b.Radius = 5
	b.Spacing = 5
	b.Label.Init(image.Rectangle{}, labelFont, size)
	b.Label.Gravity = GravityCenter
	b.Icon.Init(image.Rectangle{}, nil)
	b.Self = b
	b.StateDidChange()
}

func (b *Button) StateDidChange() {
	if b.IsDisabled() {
		b.Background = color.RGBA{R: 0xBB, G: 0xBB, B: 0xBB, A: 0xDD}
		b.Label.Color = color.Gray{0x77}
	} else if b.IsHighlighted() {
		b.Background = color.RGBA{R: 0x66, G: 0x99, B: 0xCC, A: 0xFF}
		b.Label.Color = color.White
	} else {
		b.Background = color.RGBA{R: 0xFF, G: 0xFE, B: 0xFC, A: 0xFF}
		b.Label.Color = color.Black
	}
	b.needsDisplay = true
}

func (b *Button) HandleAction(action ControlAction) {
	if fn := b.Actions[action]; fn != nil {
		fn(b)
	}
}

// Need some sort of prepare phase for drawing
func (b *Button) Draw(ctx DrawingContext) {
	b.BasicLayer.Draw(ctx)

	layout := LayoutRect{b.Rectangle.Inset(8)}

	if img := b.Icon.Image; img != nil {
		imgRect := layout.Slice(img.Bounds().Dy(), b.Spacing, FromTop).Rectangle
		b.Icon.SetFrame(imgRect)
		b.Icon.Draw(ctx)
	}

	// Render the label text in the remaining area
	b.Label.SetFrame(layout.Rectangle)
	b.Label.Draw(ctx)
}
