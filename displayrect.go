package main

import (
	"image"
	"image/color"
)

// DisplayRect projects a writable view into a subrect of the framebuffer.
// DisplayRect conforms to LayerContent
type DisplayRect struct {
	Rect
	display *Display
}

func (dr *DisplayRect) Set(x, y int, c color.Color) {
	var ok bool
	var c565 Color565
	if c565, ok = c.(Color565); !ok {
		c565 = model565.Convert(c).(Color565)
	}
	row := dr.GetRow(y)
	row[2*x], row[2*x+1] = c565.b1, c565.b2
}

func (dr *DisplayRect) ColorModel() color.Model {
	return &model565
}

func (dr *DisplayRect) Bounds() (rect image.Rectangle) {
	rect.Max.X = dr.w
	rect.Max.Y = dr.h
	return
}

func (dr *DisplayRect) At(x, y int) color.Color {
	row := dr.GetRow(y)
	return Color565{row[2*x], row[2*x+1]}
}

func (dr *DisplayRect) IsBuffered() bool {
	return false
}

func (dr *DisplayRect) GetRow(y int) []byte {
	// TODO: Bounds checking
	y += dr.y
	idx := 2 * (y*dr.display.Width + dr.x)
	return dr.display.FrameBuffer[idx : idx+2*dr.w]
}

func (dr *DisplayRect) DrawRow(row []byte, x, y int) {
	// Bounds-check before doing any real work
	if y < 0 || y >= dr.h || x >= dr.w {
		return
	}
	if x < 0 {
		row = row[-x:]
		x = 0
	}

	// For speed, duplicate GetRow() above
	y += dr.y
	bufIdx := 2 * (y*dr.display.Width + dr.x)
	bufRow := dr.display.FrameBuffer[bufIdx+2*x : bufIdx+2*dr.w]

	if len(row) > len(bufRow) {
		row = row[:len(bufRow)]
	}

	for idx, src := range row {
		if src != bufRow[idx] {
			bufRow[idx] = src
		}
	}

	// copy(bufRow, row)
}

func (dr *DisplayRect) Clip(rect Rect) DrawingContext {
	return &DisplayRect{
		Rect:    dr.Rect.Intersection(rect),
		display: dr.display,
	}
}
