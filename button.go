package touch

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

func (b *Button) ApplyColors(c ButtonColors) {
	if b.Background != c.Background {
		b.Background = c.Background
		b.Invalidate()
	}

	if b.Label.Color != c.Text {
		b.Label.Color = c.Text
		b.Label.Invalidate()
	}

	// Alpha-only images fall back to the text color when explicit Tint is not provided.
	if _, isTemplate := b.Icon.Image.(*image.Alpha); isTemplate && c.ImageTint == nil {
		c.ImageTint = c.Text
	}
	if b.Icon.Tint != c.ImageTint {
		b.Icon.Tint = c.ImageTint
		b.Icon.Invalidate()
	}
}

func (b *Button) StateDidChange() {
	if b.IsDisabled() {
		b.ApplyColors(b.Colors.Disabled)
	} else if b.IsHighlighted() {
		b.ApplyColors(b.Colors.Highlighted)
	} else {
		b.ApplyColors(b.Colors.Normal)
	}
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
		// Purposely avoid invalidation
		b.Icon.Rectangle = imgRect
	}
	b.Label.Rectangle = layout.Rectangle

	b.BasicLayer.Render(ctx)
}
