package fbui

import (
	"image"
	"image/color"
	"image/draw"
)

type ImageLayer struct {
	BasicLayer
	Image   image.Image
	Tint    color.Color
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
		if i.Tint != nil {
			// When image is an *image.Alpha, this should always hit the "happy path" in draw.drawGlyphOver
			draw.DrawMask(ctx.Image(), rect, &image.Uniform{i.Tint}, image.Point{}, i.Image, image.Point{}, draw.Over)
		} else {
			draw.Draw(ctx.Image(), rect, i.Image, image.Point{}, draw.Over)
		}
	}
}
