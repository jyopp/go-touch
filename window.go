package touch

import (
	"fmt"
	"image"
	"image/color"
	"time"
)

type Window struct {
	BufferedLayer
	display  *Display
	redrawCh chan struct{}
}

func (w *Window) Init(display *Display) {
	w.SetFrame(image.Rectangle{Max: display.Size})
	w.Self = w
	w.redrawCh = make(chan struct{}, 1)
	w.display = display
}

func (w *Window) Calibrate(ev *TouchEvent) {
	w.display.Calibration.Adjust(ev)
}

func (w *Window) InvalidateRect(rect image.Rectangle) {
	w.invalid.AddRect(rect)

	// This pattern sends a struct to the channel IFF it doesn't block.
	// Since the channel capacity is 1, this means the channel send
	// will succeed at most once per turn of the event loop
	select {
	case w.redrawCh <- struct{}{}:
	default:
	}
}

// update traverses the layer hierarchy, displaying any layers
// that need to be displayed. If any layers are displayed, a
// superset of all drawn rects is flushed to the display.
func (w *Window) update(flush func(*image.RGBA)) {
	start := time.Now()
	w.BufferedLayer.RenderBuffer()
	w.checkRoundCorners()
	drawn := time.Now()

	rects := w.dirty.Dequeue()
	for _, rect := range rects {
		flush(w.Buffer.SubImage(rect).(*image.RGBA))
	}

	if time.Since(start).Milliseconds() > 0 {
		fmt.Printf(
			"Updated: Draw %dms / Flush %dms in %v\n",
			drawn.Sub(start).Milliseconds(),
			time.Since(drawn).Milliseconds(),
			rects,
		)
	}
}

func (w *Window) checkRoundCorners() {
	// Knock off the screen corners if any of them may have been drawn over
	mask := CornerMask{w.Rectangle, w.Radius}
	// If any of the corners were drawn, mask them out before flushing
	v, h := mask.OpaqueRects()
	for _, rect := range w.dirty.Rects {
		if !(rect.In(v) || rect.In(h)) {
			// If any rect overlaps the bounds around a corner, mask them out and return
			mask.EraseCorners(w.Buffer.RGBA, color.Black)
			break
		}
	}
}

// Redraw erases the contents of the DrawBuffer and unconditonally
// redraws all layers.
// The entire DrawBuffer is flushed to the display before returning.
func (w *Window) Redraw(flush func(*image.RGBA)) {
	w.Buffer.Reset(color.RGBA{})
	w.Invalidate()
	w.update(flush)
}
