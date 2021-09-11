//go:build darwin

package touch

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework Foundation -framework QuartzCore
#import "runloop_nsapp.h"
*/
import "C"

import (
	"context"
	"image"
	"unsafe"
)

var (
	mouseEvents chan<- TouchEvent
)

//export receiveMouseEvent
func receiveMouseEvent(x, y int, pressed bool) {
	mouseEvents <- TouchEvent{
		Point:    image.Point{X: x, Y: y},
		Pressed:  pressed,
		Pressure: 0xFF,
	}
}

func mac_runloop(ctx context.Context, w *Window) {
	events := make(chan TouchEvent, 100)
	mouseEvents = events

	cW, cH := C.int(w.display.Size.X), C.int(w.display.Size.Y)

	var eventTarget LayerTouchDelegate
	var touchCanceled bool
	cancelTouch := func() {
		touchCanceled = true
		if eventTarget != nil {
			eventTarget.CancelTouch()
			eventTarget = nil
		}
	}

outer:
	for {
		select {
		case event := <-events:
			event.Cancel = cancelTouch
			if touchCanceled {
				// Ignore events until mouseup
				if !event.Pressed {
					touchCanceled = false
				}
			} else if event.Pressed {
				if eventTarget != nil {
					eventTarget.UpdateTouch(event)
				} else {
					// Only when there is no current event target, hit test for one.
					if eventTarget = w.HitTest(event); eventTarget != nil {
						eventTarget.StartTouch(event)
					}
				}
			} else {
				if eventTarget != nil {
					eventTarget.EndTouch(event)
					eventTarget = nil
				}
			}
		case task := <-runloop_tasks:
			task()
		case <-w.redrawCh:
			dirty := false
			w.update(func(_ *image.RGBA) {
				dirty = true
			})
			if dirty {
				C.DrawRGBA(unsafe.Pointer(&w.RGBA.Pix[0]), C.int(len(w.RGBA.Pix)), cW, cH)
			}
		case <-ctx.Done():
			C.StopApp()
			break outer
		}
	}
}

func _cRect(rect image.Rectangle) C.CGRect {
	size := rect.Size()
	return C.CGRectMake(C.double(rect.Min.X), C.double(rect.Min.Y), C.double(size.X), C.double(size.Y))
}

func RunLoop(ctx context.Context, w *Window) error {
	cW, cH := C.int(w.display.Size.X), C.int(w.display.Size.Y)
	C.InitApp(cW, cH)
	w.RenderBuffer()
	C.DrawRGBA(unsafe.Pointer(&w.RGBA.Pix[0]), C.int(len(w.RGBA.Pix)), cW, cH)

	go mac_runloop(ctx, w)
	C.RunApp()

	return nil
}
