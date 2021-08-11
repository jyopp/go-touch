package main

import (
	"image"
	"image/color"
)

// DisplayBuffer holds a high-color RGBA image buffer used for
// compositing before the framebuffer is written to.
// DisplayBuffer is required for compositing with transparency.
type DisplayBuffer struct {
	*image.RGBA
	Display *Display
}

func NewDisplayBuffer(display *Display, frame Rect) DisplayBuffer {
	return DisplayBuffer{
		RGBA:    image.NewRGBA(frame.Rectangle()),
		Display: display,
	}
}

// Clear sets all pixels to transparent black. Quickly.
func (b *DisplayBuffer) Clear() {
	// TODO: support clipped contexts with a separate codepath
	px := b.Pix
	l := copy(px, []byte{0x0, 0x0, 0x0, 0x0})
	// Copy l*2^n zeros on each pass
	for l < len(px) {
		copy(px[l:], px[:l])
		l *= 2
	}
}

// Set the buffer's frame. Returns true if the image data was reinitialized.
func (b *DisplayBuffer) SetFrame(frame Rect) bool {
	if frame.w == b.Rect.Dx() && frame.h == b.Rect.Dy() {
		b.Rect = frame.Rectangle()
		return false
	} else {
		b.RGBA = image.NewRGBA(frame.Rectangle())
		return true
	}
}

func (b DisplayBuffer) GetRow(y int) []byte {
	// Calculate our own pixel offset so we can truncate the row
	left := b.PixOffset(b.Rect.Min.X, y)
	right := b.PixOffset(b.Rect.Max.X, y)
	return b.Pix[left:right:right]
}

func (b DisplayBuffer) DrawRow(row []byte, x, y int) {
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

	if len(row) > len(bufRow) {
		copy(bufRow, row[:len(bufRow)])
	} else {
		copy(bufRow, row)
	}
}

// Marks a rect as needing to be drawn to the display.
// If the buffer is not associated with a display, does nothing.
func (b DisplayBuffer) SetDirty(rect Rect) {
	if b.Display == nil {
		return
	}
	b.Display.SetDirty(b.Rect.Intersect(rect.Rectangle()))
	println("Invalidated", b.Rect.String(), "~>", b.Rect.Intersect(rect.Rectangle()).String())
}

func (b DisplayBuffer) Clip(rect Rect) DrawingContext {
	// TODO: Information about rects with negative origin
	// values could be lost here, and may need special treatment.
	return DisplayBuffer{
		RGBA:    b.SubImage(rect.Rectangle()).(*image.RGBA),
		Display: b.Display,
	}
}

func (b DisplayBuffer) Fill(rect Rect, c color.Color) {
	rgba := color.RGBAModel.Convert(c).(color.RGBA)
	rowLen := rect.w * 4
	row := make([]byte, rowLen)
	for i := copy(row, []byte{rgba.R, rgba.G, rgba.B, rgba.A}); i < rowLen; {
		copy(row[i:], row[:i])
		i *= 2
	}

	// Copy the pixel row into all relevant output lines
	if rect.radius > 0 {
		// Clip the corners when drawing into an unbuffered context
		for y := 0; y < rect.h; y++ {
			i := rect.roundRectInset(y)
			b.DrawRow(row[4*i:rowLen-4*i], rect.x+i, rect.y+y)
		}
	} else {
		for y := 0; y < rect.h; y++ {
			b.DrawRow(row, rect.x, rect.y+y)
		}
	}

	b.SetDirty(rect)
}
