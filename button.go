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

var buttonFont *truetype.Font
var buttonFace font.Face
var buttonFaceOpts truetype.Options

func init() {
	var err error
	buttonFaceOpts = truetype.Options{
		Size: 14.0,
		DPI:  96.0,
	}
	if buttonFont, err = truetype.Parse(gomedium.TTF); err != nil {
		panic(err)
	}
	buttonFace = truetype.NewFace(buttonFont, &buttonFaceOpts)
}

type Button struct {
	BasicLayer
	Highlighted bool
	OnTap       func()
	Label       string
	context     *freetype.Context
	Icon        image.Image
	Disabled    bool
}

func (b *Button) Init(frame image.Rectangle) *Button {
	b.SetFrame(frame)
	b.Background = color.Transparent
	b.Radius = 5

	b.Label = "Button"
	b.context = freetype.NewContext()
	b.context.SetFont(buttonFont)
	b.context.SetFontSize(buttonFaceOpts.Size)
	b.context.SetDPI(buttonFaceOpts.DPI)

	b.Delegate = b
	return b
}

func (b *Button) SetHighlighted(highlighted bool) {
	if b.Disabled {
		highlighted = false
	}
	if highlighted == b.Highlighted {
		return
	}
	b.Highlighted = highlighted
	b.needsDisplay = true
}

// Need some sort of prepare phase for drawing
func (b *Button) Draw(layer Layer, ctx DrawingContext) {

	var bgColor, textColor color.Color
	if b.Disabled {
		bgColor = color.RGBA{R: 0xBB, G: 0xBB, B: 0xBB, A: 0xDD}
		textColor = color.Gray{0x77}
	} else if b.Highlighted {
		bgColor = color.RGBA{R: 0x66, G: 0x99, B: 0xCC, A: 0xFF}
		textColor = color.White
	} else {
		bgColor = color.RGBA{R: 0xFF, G: 0xFE, B: 0xFC, A: 0xFF}
		textColor = color.Black
	}
	ctx.Fill(b.Rectangle, bgColor, b.Radius, draw.Over)
	b.context.SetSrc(image.NewUniform(textColor))

	if b.Icon != nil {
		iconX := (b.Dx() - b.Icon.Bounds().Dx()) / 2
		offset := image.Point{-iconX, -10}
		draw.Draw(ctx.Image(), b.Rectangle, b.Icon, offset, draw.Over)
	}

	textContext := b.context
	textContext.SetDst(ctx.Image())
	textContext.SetClip(b.Rectangle)

	textWidth := font.MeasureString(buttonFace, b.Label).Round()
	textX := (b.Min.X + b.Max.X - textWidth) / 2
	if _, err := textContext.DrawString(b.Label, fixed.P(textX, b.Max.Y-13)); err != nil {
		fmt.Printf("%v drawing string: %s\n", err, b.Label)
	}
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
