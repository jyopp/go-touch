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

func (l LayoutRect) Centered(size image.Point) image.Rectangle {
	return image.Rectangle{
		Max: size,
	}.Add(l.Min.Add(l.Max).Sub(size).Div(2))
}

func (l LayoutRect) LeftCentered(size image.Point) image.Rectangle {
	rect := l.Centered(image.Point{X: l.Dx(), Y: size.Y})
	rect.Max.X = rect.Min.X + size.X
	return rect
}

func (l LayoutRect) RightCentered(size image.Point) image.Rectangle {
	rect := l.Centered(image.Point{X: l.Dx(), Y: size.Y})
	rect.Min.X = rect.Max.X - size.X
	return rect
}

func (l LayoutRect) TopLeft(size image.Point) image.Rectangle {
	return image.Rectangle{Max: size}.Add(l.Min)
}
