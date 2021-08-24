package fbui

import (
	"image"
	"image/color"
)

type ButtonColors struct {
	Background, Text, ImageTint color.Color
}

type Button struct {
	ControlLayer
	Label   TextLayer
	Icon    ImageLayer
	Actions [ControlActionsCount]func(*Button)
	Colors  struct {
		Normal, Highlighted, Disabled ButtonColors
	}

	Spacing int // Min. Distance between Icon and Label
}

func (b *Button) Init(frame image.Rectangle, labelFont string, size float64) {
	b.SetFrame(frame)
	b.Radius = 5
	b.Spacing = 5
	b.Label.Init(image.Rectangle{}, labelFont, size)
	b.Label.Gravity = GravityCenter
	b.Icon.Init(image.Rectangle{}, nil)
	b.AddChild(&b.Icon, &b.Label)
	b.Self = b

	// Default Colors
	b.Colors.Disabled.Background = color.RGBA{R: 0xBB, G: 0xBB, B: 0xBB, A: 0xDD}
	b.Colors.Disabled.Text = color.Gray{0x77}

	b.Colors.Highlighted.Background = color.RGBA{R: 0x66, G: 0x99, B: 0xCC, A: 0xCC}
	b.Colors.Highlighted.Text = color.Gray{0xFF}

	b.Colors.Normal.Background = color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	b.Colors.Normal.Text = color.Gray{0x00}

	b.StateDidChange()
}

func (b *Button) StateDidChange() {
	var colors *ButtonColors
	if b.IsDisabled() {
		colors = &b.Colors.Disabled
	} else if b.IsHighlighted() {
		colors = &b.Colors.Highlighted
	} else {
		colors = &b.Colors.Normal
	}
	b.Background = colors.Background
	b.Label.Color = colors.Text
	// Ensure that Alpha-only images are drawn with foreground color,
	// but full-color images are drawn as-is.
	if _, isTemplate := b.Icon.Image.(*image.Alpha); isTemplate {
		if colors.ImageTint != nil {
			b.Icon.Tint = colors.ImageTint
		} else {
			b.Icon.Tint = colors.Text
		}
	} else {
		b.Icon.Tint = nil
	}
	b.Invalidate()
}

func (b *Button) SetParent(parent Layer) {
	// Sort-of hack, to ensure we calculate the correct colors and states before being drawn
	b.StateDidChange()
	b.ControlLayer.SetParent(parent)
}

func (b *Button) HandleAction(action ControlAction) {
	if fn := b.Actions[action]; fn != nil {
		fn(b)
	}
}

func (b *Button) Render(ctx DrawingContext) {
	// "Layout Sublayers"
	layout := LayoutRect{b.Rectangle.Inset(8)}
	if img := b.Icon.Image; img != nil {
		imgRect := layout.Slice(img.Bounds().Dy(), b.Spacing, FromTop).Rectangle
		b.Icon.SetFrame(imgRect)
	}
	b.Label.SetFrame(layout.Rectangle)
	b.BasicLayer.Render(ctx)
}
