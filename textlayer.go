package fbui

import (
	"image"
	"image/color"
	"image/draw"
)

type TextLayer struct {
	BasicLayer
	Color    color.Color
	Gravity  image.Point
	Text     string
	textFont *Font
}

func (tl *TextLayer) Init(frame image.Rectangle, fontname string, fontsize float64) {
	tl.SetFrame(frame)
	tl.SetFont(fontname, fontsize)
	tl.Background = color.Transparent
	tl.Color = color.Black
	tl.Gravity = GravityLeft
	tl.Delegate = tl
}

func (tl *TextLayer) SetFont(name string, size float64) {
	tl.textFont = SharedFont(name, size)
}

func (tl *TextLayer) Draw(layer Layer, ctx DrawingContext) {
	if _, _, _, a := tl.Background.RGBA(); a > 0xFF {
		ctx.Fill(tl.Rectangle, tl.Background, tl.Radius, draw.Over)
	}

	layout := LayoutRect{tl.Rectangle}

	textSize := tl.textFont.Measure(tl.Text, layout.Size())
	textRect := layout.Aligned(textSize, tl.Gravity)
	tl.textFont.Draw(ctx.Image(), tl.Text, textRect, tl.Color)
}
