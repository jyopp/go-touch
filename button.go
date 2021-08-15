package main

import (
	"image"
	"image/color"
	"image/draw"
)

type Button struct {
	BasicLayer
	Highlighted bool
	OnTap       func()
	Label       TextLayer
	Icon        image.Image
	Disabled    bool
}

func (b *Button) Init(frame image.Rectangle) {
	b.SetFrame(frame)
	b.Background = color.Transparent
	b.Radius = 5
	b.Label.Init(image.Rectangle{}, systemFont, 15)
	b.Delegate = b
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

	var bgColor color.Color
	if b.Disabled {
		bgColor = color.RGBA{R: 0xBB, G: 0xBB, B: 0xBB, A: 0xDD}
		b.Label.Color = color.Gray{0x77}
	} else if b.Highlighted {
		bgColor = color.RGBA{R: 0x66, G: 0x99, B: 0xCC, A: 0xFF}
		b.Label.Color = color.White
	} else {
		bgColor = color.RGBA{R: 0xFF, G: 0xFE, B: 0xFC, A: 0xFF}
		b.Label.Color = color.Black
	}
	ctx.Fill(b.Rectangle, bgColor, b.Radius, draw.Over)

	layout := LayoutRect{b.Rectangle.Inset(8)}

	if b.Icon != nil {
		iconSize := b.Icon.Bounds().Size()
		iconBounds := layout.Slice(iconSize.Y, 5, fromTop).Aligned(iconSize, gravityCenter)
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
