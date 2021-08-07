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
