package fbui

import (
	"image"
	"image/color"
	"image/draw"
)

// DrawingContext is used for both buffered and unbuffered drawing.
// See LayerImageBuffer and DisplayRect.
type DrawingContext interface {
	Image() *image.RGBA
	Bounds() image.Rectangle

	// PutRow writes a row of pixel data into the receiver at (x, y)
	// The receiver is responsible for bounds checking:
	// x and y may be negative, and x, y, and y + len(row) may exceed receiver's dimensions
	DrawRow(row []byte, x, y int, op draw.Op)

	// Fill a rect (possibly rounded) with color.
	// Radius indicates a corner radius that will be efficiently masked.
	Fill(rect image.Rectangle, color color.Color, radius int)

	// Marks this drawing context as needing to be flushed after drawing.
	SetDirty(rect image.Rectangle)

	// Returns a DrawingContext masked to the intersection with rect.
	Clip(rect image.Rectangle) DrawingContext
}
