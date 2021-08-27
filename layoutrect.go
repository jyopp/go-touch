package touch

import "image"

type LayoutDirection int

const (
	FromLeft LayoutDirection = iota
	FromRight
	FromTop
	FromBottom
)

var (
	GravityCenter image.Point = image.Point{0x7FFF, 0x7FFF}

	GravityTop    image.Point = image.Point{0x7FFF, 0}
	GravityLeft   image.Point = image.Point{0, 0x7FFF}
	GravityBottom image.Point = image.Point{0x7FFF, 0xFFFF}
	GravityRight  image.Point = image.Point{0xFFFF, 0x7FFF}

	GravityTopLeft     image.Point = image.Point{0, 0}
	GravityTopRight    image.Point = image.Point{0xFFFF, 0}
	GravityBottomLeft  image.Point = image.Point{0, 0xFFFF}
	GravityBottomRight image.Point = image.Point{0xFFFF, 0xFFFF}
)

type LayoutRect struct {
	image.Rectangle
}

func Layout(rect image.Rectangle) (l LayoutRect) {
	l.Rectangle = rect
	return
}

func (l *LayoutRect) Slice(size, pad int, dir LayoutDirection) LayoutRect {
	sliced := *l
	switch dir {
	case FromLeft:
		sliced.Max.X = sliced.Min.X + size
		l.Min.X = sliced.Max.X + pad
	case FromRight:
		sliced.Min.X = sliced.Max.X - size
		l.Max.X = sliced.Min.X - pad
	case FromTop:
		sliced.Max.Y = sliced.Min.Y + size
		l.Min.Y = sliced.Max.Y + pad
	case FromBottom:
		sliced.Min.Y = sliced.Max.Y - size
		l.Max.Y = sliced.Min.Y - pad
	}
	return sliced
}

func (l LayoutRect) Repeat(size, pad int, dir LayoutDirection) []LayoutRect {
	rects := []LayoutRect{}
	switch dir {
	case FromLeft, FromRight:
		for l.Dx() >= size {
			rects = append(rects, l.Slice(size, pad, dir))
		}
	case FromTop, FromBottom:
		for l.Dy() >= size {
			rects = append(rects, l.Slice(size, pad, dir))
		}
	}
	return rects
}

func (l LayoutRect) Divide(count, pad int, dir LayoutDirection) []LayoutRect {
	var size int
	switch dir {
	case FromLeft, FromRight:
		size = l.Dx()
	case FromTop, FromBottom:
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

func (l LayoutRect) InsetBy(dx, dy int) LayoutRect {
	l.Min.X += dx
	l.Min.Y += dy
	l.Max.X -= dx
	l.Max.Y -= dy
	return l
}
