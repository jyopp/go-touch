package fbui

import (
	"fmt"
	"image"
	"image/color"
	"time"
)

// This file contains methods for invalidating and drawing views.

type Window struct {
	BufferedLayer
	display *Display
}

func (w *Window) Init(display *Display) {
	w.SetFrame(display.Bounds())
	w.Self = w
	w.display = display
}

func (w *Window) Calibrate(ev *TouchEvent) {
	w.display.Calibration.Adjust(ev)
}

// update traverses the layer hierarchy, displaying any layers
// that need to be displayed. If any layers are displayed, a
// superset of all drawn rects is flushed to the display.
func (w *Window) update() {
	start := time.Now()
	w.BufferedLayer.DisplayIfNeeded(nil)
	w.checkRoundCorners()
	drawn := time.Now()

	if w.dirty.Reduce() == 0 {
		return
	}

	for _, rect := range w.dirty.Rects {
		w.display.render(w.Buffer.SubImage(rect).(*image.RGBA))
	}

	if time.Since(start).Milliseconds() > 0 {
		fmt.Printf(
			"Updated: Draw %dms / Flush %dms in %v\n",
			drawn.Sub(start).Milliseconds(),
			time.Since(drawn).Milliseconds(),
			w.dirty.Rects,
		)
	}
	// Truncate without modifying underlying storage or capacity
	w.dirty.Clear()
}

func (w *Window) checkRoundCorners() {
	// Knock off the screen corners if any of them may have been drawn over
	mask := CornerMask{w.Rectangle, 9}
	// If any of the corners were drawn, mask them out before flushing
	v, h := mask.OpaqueRects()
	for _, rect := range w.dirty.Rects {
		if !(rect.In(v) || rect.In(h)) {
			// If any rect overlaps the bounds around a corner, mask them out and return
			mask.EraseCorners(w.Buffer.RGBA)
			break
		}
	}
}

func (w *Window) SetDirty(rect image.Rectangle) {
	w.dirty.AddRect(rect)
}

// Redraw erases the contents of the DrawBuffer and unconditonally
// redraws all layers.
// The entire DrawBuffer is flushed to the display before returning.
func (w *Window) Redraw() {
	w.Buffer.Reset(color.RGBA{})
	w.BufferedLayer.Display(nil)
	w.display.render(w.Buffer.RGBA)
	w.dirty.Clear()
}
