package main

import (
	"image"
	"image/color"
	"image/draw"
)

type ImageLayer struct {
	BasicLayer
	Image      image.Image
	Background color.Color
	Centered   bool
}

func (i *ImageLayer) Init(frame image.Rectangle, image image.Image) *ImageLayer {
	i.Image = image
	i.Centered = true
	i.BasicLayer.Init(frame, i)
	i.Background = color.Transparent
	return i
}

func (i *ImageLayer) Draw(layer Layer, ctx DrawingContext) {
	if _, _, _, a := i.Background.RGBA(); a > 0x10 {
		ctx.Fill(i.Frame(), i.Background, i.radius, draw.Over)
	}

	if i.Image != nil {
		var negativeOffset image.Point
		if i.Centered {
			negativeOffset.X = (i.Image.Bounds().Dx() - i.Dx()) / 2
			negativeOffset.Y = (i.Image.Bounds().Dy() - i.Dy()) / 2
		}
		draw.Draw(ctx, ctx.Bounds(), i.Image, negativeOffset, draw.Over)
	}
}
