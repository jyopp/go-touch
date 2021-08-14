package main

import (
	"image"
	"image/color"
	"image/draw"
)

// DisplayBuffer holds a high-color RGBA image buffer used for
// compositing before the framebuffer is written to.
// DisplayBuffer is required for compositing with transparency.
type DisplayBuffer struct {
	*image.RGBA
	Display   *Display
	isClipped bool
}

func NewDisplayBuffer(display *Display, frame image.Rectangle) *DisplayBuffer {
	return &DisplayBuffer{
		RGBA:    image.NewRGBA(frame),
		Display: display,
	}
}

func (b *DisplayBuffer) Image() *image.RGBA {
	return b.RGBA
}

func (b *DisplayBuffer) Bounds() image.Rectangle {
	return b.Rect
}

// Reset resets every pixel in the buffer as efficiently as possible.
func (b *DisplayBuffer) Reset(c color.Color) {
	// TODO: support clipped contexts with a separate codepath
	px := b.Pix
	rgba := color.RGBAModel.Convert(c).(color.RGBA)
	l := copy(px, []byte{rgba.R, rgba.G, rgba.B, rgba.A})
	// Copy l*2^n zeros on each pass
	for ; l < len(px); l *= 2 {
		copy(px[l:], px[:l])
	}
}

// Set the buffer's frame. Returns true if the image data was reinitialized.
func (b *DisplayBuffer) SetFrame(frame image.Rectangle) bool {
	if frame.Size().Eq(b.Rect.Size()) {
		b.Rect = frame
		return false
	} else {
		b.RGBA = image.NewRGBA(frame)
		return true
	}
}

func (b *DisplayBuffer) GetRow(y int) []byte {
	// Calculate our own pixel offset so we can truncate the row
	left := b.PixOffset(b.Rect.Min.X, y)
	right := b.PixOffset(b.Rect.Max.X, y)
	return b.Pix[left:right:right]
}

func (b *DisplayBuffer) DrawRow(row []byte, x, y int, op draw.Op) {
	// Bounds-check and adjust before copying pixel data
	min, max := b.Rect.Min, b.Rect.Max
	if y < min.Y {
		return
	}

	if x < min.X {
		row = row[-4*x:]
		x = 0
	}

	if x >= max.X || y >= max.Y {
		// Origin is beyond extents, no-op
		return
	}

	// Calculate our own pixel offset so we can truncate the row
	bufRow := b.Pix[b.PixOffset(x, y):b.PixOffset(max.X, y)]

	if len(bufRow) < len(row) {
		row = row[:len(bufRow)]
	}

	if op == draw.Src {
		copy(bufRow, row)
		return
	}

	// Else drawing mode is 'Over' and we may need to read pixels
	for i := 0; i < len(row); i += 4 {
		sPxl := row[i : i+4 : i+4]
		dPxl := bufRow[i : i+4 : i+4]
		sA := sPxl[3]
		if sA == 0xFF {
			dPxl[0] = sPxl[0]
			dPxl[1] = sPxl[1]
			dPxl[2] = sPxl[2]
			dPxl[3] = sPxl[3]
		} else {
			// Source alpha is premultiplied, get its inverse for blending.
			dA := uint32(^sA)
			dPxl[0] = sPxl[0] + byte((dA*uint32(dPxl[0]))>>8)
			dPxl[1] = sPxl[1] + byte((dA*uint32(dPxl[1]))>>8)
			dPxl[2] = sPxl[2] + byte((dA*uint32(dPxl[2]))>>8)
			dPxl[3] = sA + byte((dA*uint32(dPxl[3]))>>8)
		}
	}
}

// Marks a rect as needing to be drawn to the display.
// If the buffer is not associated with a display, does nothing.
func (b *DisplayBuffer) SetDirty(rect image.Rectangle) {
	if b.Display != nil {
		b.Display.SetDirty(rect.Intersect(b.Rect))
	}
}

func (b *DisplayBuffer) Clip(rect image.Rectangle) DrawingContext {
	rect = rect.Intersect(b.Rect)
	if rect.Empty() {
		return nil
	}

	// TODO: Information about rects with negative origin
	// values could be lost here, and may need special treatment.
	return &DisplayBuffer{
		RGBA:      b.SubImage(rect).(*image.RGBA),
		Display:   b.Display,
		isClipped: true,
	}
}

func (b *DisplayBuffer) Fill(rect image.Rectangle, c color.Color, radius int, op draw.Op) {
	mask := CornerMask{rect, radius}
	if !b.isClipped && b.Rect.In(rect) {
		// Fastest path;
		b.Reset(c)
		mask.EraseCorners(b)
		return
	}

	rgba := color.RGBAModel.Convert(c).(color.RGBA)
	rowLen := rect.Dx() * 4
	row := make([]byte, rowLen)
	for i := copy(row, []byte{rgba.R, rgba.G, rgba.B, rgba.A}); i < rowLen; {
		copy(row[i:], row[:i])
		i *= 2
	}

	if rgba.A < 0xFF {
		op = draw.Over
	}

	// Copy the pixel row into all relevant output lines
	if radius > 0 {
		// Clip the corners when drawing into an unbuffered context
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			i := mask.RowInset(y)
			b.DrawRow(row[4*i:rowLen-4*i], rect.Min.X+i, y, op)
		}
	} else {
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			b.DrawRow(row, rect.Min.X, y, op)
		}
	}

	b.SetDirty(rect)
}
