package fbui

import (
	"image"
	"image/color"
	"image/draw"
)

// RenderedText caches the alphamask and dimensions of a text string
type RenderedText struct {
	Text     string
	MaxSize  image.Point
	font     *Font
	rendered *image.Alpha
}

func (rt *RenderedText) Invalidate() {
	rt.rendered = nil
}

func (rt *RenderedText) SetFont(name string, size float64) {
	f := SharedFont(name, size)
	if f == rt.font {
		return
	}
	rt.font = f
	rt.rendered = nil
}

func (rt *RenderedText) Render() {
	rect := image.Rectangle{
		Max: rt.font.Measure(rt.Text, rt.MaxSize),
	}
	rt.rendered = image.NewAlpha(rect)
	rt.font.Draw(rt.rendered, rt.Text, rect, color.Alpha{0xFF})
}

// Prepare creates or updates the cached bitmap for the given string.
// Returns the size of the rendered image.
func (rt *RenderedText) Prepare(text string, maxSize image.Point) image.Point {
	if rt.rendered == nil || !maxSize.Eq(rt.MaxSize) || text != rt.Text {
		rt.MaxSize = maxSize
		rt.Text = text
		rt.Render()
	}
	return rt.rendered.Rect.Size()
}

// Draw draws the text into the given image/rect.
// If Render() or Prepare(...) have not been called since the last
// invalidation, a panic may result.
func (rt *RenderedText) Draw(img draw.Image, rect image.Rectangle, c color.Color) {
	colorSrc := image.NewUniform(c)
	draw.DrawMask(img, rect, colorSrc, image.Point{}, rt.rendered, image.Point{}, draw.Over)
}
