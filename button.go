package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/math/fixed"
)

var boldFont *truetype.Font

func init() {
	var err error
	if boldFont, err = truetype.Parse(gobold.TTF); err != nil {
		panic(err)
	}
}

type Button struct {
	*Layer
	Highlighted bool
	OnTap       func()
	Label       string
	context     *freetype.Context
	// textImage   *image.Gray16
}

func NewButton(r Rect) *Button {
	button := &Button{
		Layer: NewLayer(r),
	}
	button.Layer.Owner = button
	button.rounded = true
	button.NeedsDisplay = true

	// button.textImage = image.NewGray16(button.Layer.Bounds())

	button.Label = "Button"
	button.context = freetype.NewContext()
	button.context.SetDPI(72.0)
	button.context.SetFont(boldFont)
	button.context.SetFontSize(26.0)
	button.context.SetDst(button.Layer)
	button.context.SetClip(button.Layer.Bounds())
	button.context.SetSrc(image.NewUniform(color.Black))

	return button
}

func (b *Button) SetHighlighted(highlighted bool) {
	if highlighted == b.Highlighted {
		return
	}
	b.Highlighted = highlighted
	b.DrawLayer()
}

func (b *Button) DrawLayer() {
	if b.Highlighted {
		b.FillRGB(0x55, 0xAA, 0xCC)
	} else {
		b.FillRGB(0xFF, 0xFE, 0xFC)
	}

	if _, err := b.context.DrawString(b.Label, fixed.P(60, int(b.h/2)+10)); err != nil {
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
