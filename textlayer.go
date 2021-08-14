package main

import (
	"image"
	"image/color"
	"image/draw"
)

type TextLayer struct {
	BasicLayer
	Color     color.Color
	Padding   int
	Text      string
	drawnText RenderedText
}

func (tl *TextLayer) Init(frame image.Rectangle, fontname string, fontsize float64) {
	tl.SetFrame(frame)
	tl.SetFont(fontname, fontsize)
	tl.Background = color.Transparent
	tl.Color = color.Black
	tl.Delegate = tl
}

func (tl *TextLayer) SetFont(name string, size float64) {
	tl.drawnText.SetFont(name, size)
}

func (tl *TextLayer) Draw(layer Layer, ctx DrawingContext) {
	if _, _, _, a := tl.Background.RGBA(); a > 0xFF {
		ctx.Fill(tl.Rectangle, tl.Background, tl.Radius, draw.Over)
	}

	layout := LayoutRect{tl.Rectangle.Inset(tl.Padding)}

	textMask := tl.drawnText.Prepare(tl.Text, tl.Size())
	textRect := layout.LeftCentered(textMask.Rect.Size())
	textSrc := image.NewUniform(tl.Color)
	zp := image.Point{}
	draw.DrawMask(ctx.Image(), textRect, textSrc, zp, textMask, zp, draw.Over)
}
