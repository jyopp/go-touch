package touch

import (
	"image"
	"image/color"
)

type TextLayer struct {
	BasicLayer
	Color    color.Color
	Gravity  image.Point
	Padding  image.Point
	Text     string
	textFont *Font
}

func (tl *TextLayer) Init(frame image.Rectangle, fontname string, fontsize float64) {
	tl.SetFrame(frame)
	tl.SetFont(fontname, fontsize)
	tl.Color = color.Black
	tl.Gravity = GravityLeft
	tl.Self = tl
}

func (tl *TextLayer) SetFont(name string, size float64) {
	tl.textFont = SharedFont(name, size)
	tl.Invalidate()
}

func (tl *TextLayer) SetText(text string) {
	tl.Text = text
	tl.Invalidate()
}

func (tl *TextLayer) NaturalSize() image.Point {
	return tl.textFont.Measure(tl.Text).Add(tl.Padding.Mul(2))
}

func (tl *TextLayer) DrawIn(ctx DrawingContext) {
	tl.BasicLayer.DrawIn(ctx)

	layout := Layout(tl.Rectangle).InsetBy(tl.Padding.X, tl.Padding.Y)

	textSize := tl.textFont.MeasureIn(tl.Text, layout.Size())
	textRect := layout.Aligned(textSize, tl.Gravity)
	tl.textFont.Draw(ctx.Image(), tl.Text, textRect, tl.Color)
}
