package main

import (
	"image/color"
	"image/draw"
)

// DrawingContext is used for both buffered and unbuffered drawing.
// See LayerImageBuffer and DisplayRect.
type DrawingContext interface {
	draw.Image
	// GetRow returns a slice of bytes with pixel data for one line of the raster.
	GetRow(y int) []byte
	// PutRow writes a row of pixel data into the receiver at (x, y)
	// The receiver is responsible for bounds checking:
	// x and y may be negative, and x, y, and y + len(row) may exceed receiver's dimensions
	DrawRow(row []byte, x, y int, op draw.Op)

	// Fill a rect (possibly rounded) with color
	Fill(rect Rect, color color.Color)

	// Marks this drawing context as needing to be flushed after drawing.
	SetDirty(rect Rect)

	// Returns a DrawingContext masked to the intersection with rect.
	Clip(rect Rect) DrawingContext
}
