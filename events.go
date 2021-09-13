package touch

import (
	"image"
)

type TouchEvent struct {
	image.Point
	Pressed  bool
	Pressure int
	Cancel   func()
}

func (e TouchEvent) InRadius(e2 TouchEvent, r int) bool {
	x, y := e.X-e2.X, e.Y-e2.Y
	return x*x+y*y <= r*r
}
