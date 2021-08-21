package fbui

import (
	"image"
	"image/draw"
)

type ImageLayer struct {
	BasicLayer
	Image   image.Image
	Gravity image.Point
}

func (i *ImageLayer) Init(frame image.Rectangle, image image.Image) {
	i.SetFrame(frame)
	i.Image = image
	i.Gravity = GravityCenter
	i.Self = i
}

func (i *ImageLayer) DrawIn(ctx DrawingContext) {
	i.BasicLayer.DrawIn(ctx)

	if i.Image != nil {
		size := i.Image.Bounds().Size()
		rect := Layout(i.Rectangle).Aligned(size, i.Gravity)
		draw.Draw(ctx.Image(), rect, i.Image, image.Point{}, draw.Over)
	}
}
