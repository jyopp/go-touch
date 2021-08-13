package main

import "image"

type Rect struct {
	x, y int
	w, h int
	// Valid values are 0-7
	radius int
}

func (r Rect) Bounds() Rect {
	return Rect{0, 0, r.w, r.h, r.radius}
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

func (r Rect) Intersection(r2 Rect) (rect Rect) {
	if r.x > r2.x {
		rect.x = r.x
	} else {
		rect.x = r2.x
	}
	r1r, r2r := r.x+r.w, r2.x+r2.w
	if r1r < r2r {
		rect.w = r1r - rect.x
	} else {
		rect.w = r2r - rect.x
	}

	if r.y > r2.y {
		rect.y = r.y
	} else {
		rect.y = r2.y
	}
	r1b, r2b := r.y+r.h, r2.y+r2.h
	if r1b < r2b {
		rect.h = r1b - rect.y
	} else {
		rect.h = r2b - rect.y
	}

	rect.radius = r2.radius

	return
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
