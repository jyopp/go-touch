package main

type Rect struct {
	x, y    int32
	w, h    int32
	rounded bool
}

func (r Rect) Bottom() int32 {
	return r.y + r.h
}

func (r Rect) Right() int32 {
	return r.x + r.w
}

func (r Rect) Contains(x, y int32) bool {
	return x >= r.x && y >= r.y && x < r.x+r.w && y < r.y+r.h
}

func (r Rect) Inset(dx, dy int32) Rect {
	return Rect{
		x: r.x + dx,
		y: r.y + dy,
		w: r.w - 2*dx,
		h: r.h - 2*dy,
	}
}

func (r *Rect) SliceV(y, pad int32) Rect {
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

func (r *Rect) GridVHeight(y, pad int32) []Rect {
	remain := *r
	rects := []Rect{}
	for remain.h >= y {
		rects = append(rects, remain.SliceV(y, pad))
	}
	return rects
}

func (r *Rect) GridVCount(count, pad int32) []Rect {
	itemH := (r.h + pad) / count
	return r.GridVHeight(itemH-pad, pad)
}

var _roundInsets = [5]int32{5, 3, 2, 1, 1}

// private function; Returns the number of pixels that should be clipped from a given line
func (r Rect) roundRectInset(line int32) int32 {
	if line < 5 {
		return _roundInsets[line]
	}
	if line > r.h-6 {
		return _roundInsets[r.h-line-1]
	}
	return 0
}
