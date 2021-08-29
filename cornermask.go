package touch

import (
	"image"
	"image/color"
	"image/draw"
)

// CornerMask is a read-only image.Image with a dynamically
// calculated alpha channel, and no internal bitmap storage.
type CornerMask struct {
	image.Rectangle
	Radius int
}

// Returns a mask image with uniform alpha value, with corners
// removed if appropriate
func (cm CornerMask) AlphaImage(opacity byte) *image.Alpha {
	r := cm.Radius
	if cm.Rectangle.Empty() || ((r == 0 || r > MaxCornerRadius) && opacity == 0xFF) {
		return nil
	}

	alpha := image.NewAlpha(cm.Rectangle)
	bytesFill(alpha.Pix, []byte{opacity})
	cm.EraseCorners(alpha, color.Alpha{0x00})
	return alpha
}

// OpaqueRects returns a vertical and horizontal covering rectangle
// that, when combined, cover the opaque (non-corner) parts of the
// masked rectangle.
func (cm CornerMask) OpaqueRects() (v, h image.Rectangle) {
	v, h = cm.Rectangle, cm.Rectangle
	v.Min.X += cm.Radius
	v.Max.X -= cm.Radius
	h.Min.Y += cm.Radius
	h.Max.Y -= cm.Radius
	return
}

func (cm CornerMask) EraseCorners(img draw.Image, c color.Color) {
	c = img.ColorModel().Convert(c)
	r := cm.Radius
	if r == 0 {
		return
	} else if r > MaxCornerRadius {
		// Draw something awful to highlight unsupported values
		c = img.ColorModel().Convert(
			color.RGBA{R: 0, G: 0x80, B: 0, A: 0x80},
		)
		r = MaxCornerRadius
	}

	min, max := cm.Min, cm.Max
	for row, inset := range _roundInsets[r] {
		yTop, yBottom := min.Y+row, max.Y-row-1
		for col := 0; col < inset; col++ {
			xLeft, xRight := min.X+col, max.X-col-1
			img.Set(xLeft, yTop, c)
			img.Set(xRight, yTop, c)
			img.Set(xLeft, yBottom, c)
			img.Set(xRight, yBottom, c)
		}
	}
}

func (cm CornerMask) RowInset(y int) int {
	r := cm.Radius
	y -= cm.Min.Y
	if y >= r {
		y = cm.Dy() - y - 1
	}
	if r > MaxCornerRadius || y < 0 || y >= r {
		return 0
	}
	return _roundInsets[r][y]
}

const MaxCornerRadius = 9

// This format can be visualized as the number of perpendicular
// pixels that should be erased, given an x or y inset from the edge
var _roundInsets = [MaxCornerRadius + 1][]int{
	{},
	{1},
	{2, 1},
	{3, 2, 1},
	{4, 2, 1, 1},
	{5, 3, 2, 1, 1},
	{6, 4, 3, 2, 1, 1},
	{7, 5, 3, 2, 2, 1, 1},
	{8, 6, 4, 3, 2, 2, 1, 1},
	{9, 6, 4, 3, 2, 2, 1, 1, 1},
}
