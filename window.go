package fbui

import (
	"fmt"
	"image"
	"image/color"
	"sort"
	"time"
)

// This file contains methods for invalidating and drawing views.

type Window struct {
	BufferedLayer
	display    *Display
	dirtyRects []image.Rectangle
}

func (w *Window) Init(display *Display) {
	w.SetFrame(display.Bounds())
	w.Buffer.ctx = w
	w.Self = w
	w.display = display
}

func (w *Window) Calibrate(ev *TouchEvent) {
	w.display.Calibration.Adjust(ev)
}

// TODO: This can be made much more complex, returning an array-of-rects
// For example if r1.Intersect(r2).(Min,Max).Y == r2.(Min,Max).Y, the
// intersecting part of r2 should be removed and the exclusion returned.
func shouldMergeDrawRects(r1, r2 image.Rectangle) bool {
	if !r1.Overlaps(r2) {
		return false
	} else if r1.Min.Y <= r2.Min.Y {
		// r1.Min is above or at r2.Min
		return r2.Max.Y <= r1.Min.Y
	} else {
		return r2.Max.Y >= r1.Max.Y
	}
}

// update traverses the layer hierarchy, displaying any layers
// that need to be displayed. If any layers are displayed, a
// superset of all drawn rects is flushed to the display.
func (w *Window) update() {
	start := time.Now()
	w.BufferedLayer.DisplayIfNeeded(nil)
	w.checkRoundCorners()
	drawn := time.Now()

	// Don't delegate to Flush because we're dumping diagnostic logs...
	w.mergeDirtyRects()
	for _, rect := range w.dirtyRects {
		w.display.flush(w.Buffer.SubImage(rect).(*image.RGBA))
	}

	if time.Since(start).Milliseconds() > 0 {
		fmt.Printf(
			"Updated: Draw %dms / Flush %dms in %v\n",
			drawn.Sub(start).Milliseconds(),
			time.Since(drawn).Milliseconds(),
			w.dirtyRects,
		)
	}
	// Truncate without modifying underlying storage or capacity
	w.dirtyRects = w.dirtyRects[:0]
}

func (w *Window) checkRoundCorners() {
	// Knock off the screen corners if any of them may have been drawn over
	mask := CornerMask{w.Rectangle, 9}
	// If any of the corners were drawn, mask them out before flushing
	v, h := mask.OpaqueRects()
	for _, rect := range w.dirtyRects {
		if rect.In(v) || rect.In(h) {
			continue
		}
		mask.EraseCorners(w.Buffer.RGBA)
	}
}

func (w *Window) mergeDirtyRects() {
	rects := w.dirtyRects
	if len(rects) == 0 {
		return
	}
	// Sort rects by their Min.Y in ascending order
	sort.Slice(rects, func(i, j int) bool {
		return rects[i].Min.Y < rects[j].Min.Y
	})
	// Reduce all overlapping rects.
	// This is a heuristic, vulnerable to some worst-case patterns.
	rect1 := rects[0]
	for idx := 1; idx < len(rects); idx++ {
		rect2 := rects[idx]
		if rect1.Overlaps(rect2) {
			rect1 = rect1.Union(rect2)
			rects[idx] = rect1
			rects[idx-1] = image.Rectangle{}
		} else {
			rect1 = rects[idx]
		}
	}
	sort.Slice(rects, func(i, j int) bool {
		rI, rJ := rects[i], rects[j]
		if rI.Empty() {
			return false
		} else if rJ.Empty() {
			return true
		}
		return rects[i].Min.Y < rects[j].Min.Y
	})
	// Truncate all the empty rects out.
	i := len(rects)
	for i > 0 && rects[i-1].Empty() {
		i--
	}
	w.dirtyRects = rects[:i]
}

// SetDirty expands or appends a dirty rect to include all pixels in rect.
func (w *Window) SetDirty(rect image.Rectangle) {
	for idx := range w.dirtyRects {
		if shouldMergeDrawRects(rect, w.dirtyRects[idx]) {
			w.dirtyRects[idx] = rect.Union(w.dirtyRects[idx])
			return
		}
	}
	w.dirtyRects = append(w.dirtyRects, rect)
}

// Redraw erases the contents of the DrawBuffer and unconditonally
// redraws all layers.
// The entire DrawBuffer is flushed to the display before returning.
func (w *Window) Redraw() {
	w.Buffer.Reset(color.RGBA{})
	w.BufferedLayer.Display(nil)
	w.display.flush(w.Buffer.RGBA)
	w.dirtyRects = w.dirtyRects[:0]
}
