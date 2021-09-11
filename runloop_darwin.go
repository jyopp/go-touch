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
	"runtime"
	"unsafe"
)

func init() {
	// Necessary to ensure that main() is called from the correct thread
	runtime.LockOSThread()
}

//export receiveMouseEvent
func receiveMouseEvent(x, y int, pressed bool) {
	MainRunLoop.events <- TouchEvent{
		Point:    image.Point{X: x, Y: y},
		Pressed:  pressed,
		Pressure: 0xFF,
	}
}

func (runloop *RunLoop) platformInit() {
	runloop.events = make(chan TouchEvent, 100)

	window := runloop.Window
	// Don't allow round corners for windowed display
	window.Radius = 0

	cW, cH := C.int(window.display.Size.X), C.int(window.display.Size.Y)
	C.InitApp(cW, cH)
}

func (runloop *RunLoop) updateDisplay() {
	win := runloop.Window
	cW, cH := C.int(win.display.Size.X), C.int(win.display.Size.Y)

	dirty := false
	win.update(func(_ *image.RGBA) {
		dirty = true
	})
	if dirty {
		C.DrawRGBA(unsafe.Pointer(&win.RGBA.Pix[0]), C.int(len(win.RGBA.Pix)), cW, cH)
	}
}

func (runloop *RunLoop) cleanup() {
	C.StopApp()
}

func (runloop *RunLoop) Run(ctx context.Context) {
	go runloop.runInner(ctx)
	C.RunApp()
}
