package main

import "image"

type Rect struct {
	x, y int
	w, h int
	// Valid values are 0-7
	radius int
}

func (r Rect) ContentBounds() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{r.w, r.h},
	}
}

func (r Rect) Rectangle() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{r.x, r.y},
		Max: image.Point{r.x + r.w, r.y + r.h},
	}
}

func (r Rect) Bottom() int {
	return r.y + r.h
}

func (r Rect) Right() int {
	return r.x + r.w
}

func (r Rect) Contains(x, y int) bool {
	return x >= r.x && y >= r.y && x < r.x+r.w && y < r.y+r.h
}

func (r Rect) Inset(dx, dy int) Rect {
	return Rect{
		x: r.x + dx,
		y: r.y + dy,
		w: r.w - 2*dx,
		h: r.h - 2*dy,
	}
}

func (r *Rect) SliceV(y, pad int) Rect {
	sliced := *r
	if y < 0 {
		sliced.h = -y
		sliced.y = r.Bottom() + y
		r.h -= (pad - y)
	} else {
		sliced.h = y
		dY := y + pad
		r.y += dY
		r.h -= dY
	}
	return sliced
}

func (r *Rect) GridVHeight(y, pad int) []Rect {
	remain := *r
	rects := []Rect{}
	for remain.h >= y {
		rects = append(rects, remain.SliceV(y, pad))
	}
	return rects
}

func (r *Rect) GridVCount(count, pad int) []Rect {
	itemH := (r.h + pad) / count
	return r.GridVHeight(itemH-pad, pad)
}

func (r *Rect) SliceH(x, pad int) Rect {
	sliced := *r
	if x < 0 {
		sliced.w = -x
		sliced.x = r.Right() + x
		r.w -= (pad - x)
	} else {
		sliced.w = x
		dX := x + pad
		r.x += dX
		r.w -= dX
	}
	return sliced
}

func (r *Rect) GridHWidth(x, pad int) []Rect {
	remain := *r
	rects := []Rect{}
	for remain.w >= x {
		rects = append(rects, remain.SliceH(x, pad))
	}
	return rects
}

func (r *Rect) GridHCount(count, pad int) []Rect {
	itemW := (r.w + pad) / count
	return r.GridHWidth(itemW-pad, pad)
}

var _roundInsets = [9][]int{
	{},
	{1},
	{2, 1},
	{3, 2, 1},
	{4, 2, 1, 1},
	{5, 3, 2, 1, 1},
	{6, 4, 3, 2, 1, 1},
	{7, 5, 3, 2, 2, 1, 1},
	{8, 6, 4, 3, 2, 2, 1, 1},
}

// private function; Returns the number of pixels that should be clipped from a given line
func (r Rect) roundRectInset(line int) int {
	if r.radius < 9 {
		if line < r.radius {
			return _roundInsets[r.radius][line]
		}
		if line >= r.h-r.radius {
			return _roundInsets[r.radius][r.h-line-1]
		}
	}
	return 0
}
