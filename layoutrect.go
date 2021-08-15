package main

import "image"

type LayoutRect struct {
	image.Rectangle
}

type LayoutDirection int

const (
	fromLeft LayoutDirection = iota
	fromRight
	fromTop
	fromBottom
)

var (
	gravityCenter image.Point = image.Point{0x7FFF, 0x7FFF}

	gravityTop    image.Point = image.Point{0x7FFF, 0}
	gravityLeft   image.Point = image.Point{0, 0x7FFF}
	gravityBottom image.Point = image.Point{0x7FFF, 0xFFFF}
	gravityRight  image.Point = image.Point{0xFFFF, 0x7FFF}

	gravityTopLeft     image.Point = image.Point{0, 0}
	gravityTopRight    image.Point = image.Point{0xFFFF, 0}
	gravityBottomLeft  image.Point = image.Point{0, 0xFFFF}
	gravityBottomRight image.Point = image.Point{0xFFFF, 0xFFFF}
)

func (l *LayoutRect) Slice(size, pad int, dir LayoutDirection) LayoutRect {
	sliced := *l
	switch dir {
	case fromLeft:
		sliced.Max.X = sliced.Min.X + size
		l.Min.X = sliced.Max.X + pad
	case fromRight:
		sliced.Min.X = sliced.Max.X - size
		l.Max.X = sliced.Min.X - pad
	case fromTop:
		sliced.Max.Y = sliced.Min.Y + size
		l.Min.Y = sliced.Max.Y + pad
	case fromBottom:
		sliced.Min.Y = sliced.Max.Y - size
		l.Max.Y = sliced.Min.Y - pad
	}
	return sliced
}

func (l *LayoutRect) Repeat(size, pad int, dir LayoutDirection) []LayoutRect {
	remain := *l
	rects := []LayoutRect{}
	switch dir {
	case fromLeft, fromRight:
		for remain.Dx() >= size {
			rects = append(rects, remain.Slice(size, pad, dir))
		}
	case fromTop, fromBottom:
		for remain.Dy() >= size {
			rects = append(rects, remain.Slice(size, pad, dir))
		}
	}
	return rects
}

func (l *LayoutRect) Divide(count, pad int, dir LayoutDirection) []LayoutRect {
	var size int
	switch dir {
	case fromLeft, fromRight:
		size = l.Dx()
	case fromTop, fromBottom:
		size = l.Dy()
	}
	size = (size+pad)/count - pad
	return l.Repeat(size, pad, dir)
}

// Aligned returns a rect with the given size, aligned such that
// the unit point at "gravity" in the layout rect and the returned
// rect are coincident.
// Gravity is expressed as a unit value from 0-65535.
//   lr.Aligned(size, image.Point{0, 32768}) // Rect at left-center
func (l LayoutRect) Aligned(size, gravity image.Point) image.Rectangle {
	// Smoothly scale between (origin at Min) and (origin at Max - size)
	gp := image.Point{
		X: l.Min.X*(0xFFFF-gravity.X) + (l.Max.X-size.X)*gravity.X,
		Y: l.Min.Y*(0xFFFF-gravity.Y) + (l.Max.Y-size.Y)*gravity.Y,
	}.Div(0xFFFF)

	return image.Rectangle{
		Min: gp,
		Max: gp.Add(size),
	}
}
