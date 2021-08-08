package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomedium"
	"golang.org/x/image/math/fixed"
)

const buttonFontSize = 21.0

var buttonFont *truetype.Font
var buttonFace font.Face

func init() {
	var err error
	if buttonFont, err = truetype.Parse(gomedium.TTF); err != nil {
		panic(err)
	}
	buttonFace = truetype.NewFace(buttonFont, &truetype.Options{Size: buttonFontSize})
}

type Button struct {
	BasicLayer
	Highlighted bool
	OnTap       func()
	Label       string
	context     *freetype.Context
	Icon        image.Image
}

func NewButton(r Rect) *Button {
	button := &Button{}
	button.BasicLayer = *NewLayer(r, button)
	button.radius = 5

	button.Label = "Button"
	button.context = freetype.NewContext()
	button.context.SetDPI(72.0)
	button.context.SetFont(buttonFont)
	button.context.SetFontSize(buttonFontSize)
	button.context.SetSrc(image.NewUniform(color.Black))

	return button
}

func (b *Button) SetHighlighted(highlighted bool) {
	if highlighted == b.Highlighted {
		return
	}
	b.Highlighted = highlighted
	b.needsDisplay = true
}

func (b *Button) Draw(layer Layer, ctx LayerDrawing) {

	if b.Highlighted {
		model565.FillRGB(ctx, b.Rect, 0x66, 0x99, 0xCC)
		b.context.SetSrc(image.NewUniform(color.White))
	} else {
		model565.FillRGB(ctx, b.Rect, 0xFF, 0xFE, 0xFC)
		b.context.SetSrc(image.NewUniform(color.Black))
	}

	if b.Icon != nil {
		iconX := (b.w - b.Icon.Bounds().Dx()) / 2
		draw.Draw(ctx, layer.Frame().Rectangle(), b.Icon, image.Pt(-iconX, -3), draw.Src)
	}

	textContext := b.context
	textContext.SetDst(ctx)
	textContext.SetClip(layer.Frame().Rectangle())

	textWidth := font.MeasureString(buttonFace, b.Label).Round()
	textX := b.x + (b.w-textWidth)/2
	if _, err := textContext.DrawString(b.Label, fixed.P(textX, b.Bottom()-10)); err != nil {
		fmt.Printf("%v drawing string: %s\n", err, b.Label)
	}
}

func (b *Button) StartTouch(event TouchEvent) {
	b.SetHighlighted(b.Contains(event.X, event.Y))
}

func (b *Button) UpdateTouch(event TouchEvent) {
	b.SetHighlighted(b.Contains(event.X, event.Y))
}

func (b *Button) EndTouch(event TouchEvent) {
	if b.Contains(event.X, event.Y) && b.OnTap != nil {
		b.OnTap()
	}
	b.SetHighlighted(false)
}
