//go:build !linux

package touch

import (
	"context"
	"image"
	"os"
	"path/filepath"

	"github.com/tfriedel6/canvas/sdlcanvas"
	"golang.org/x/image/draw"
)

func bindMouseEvents(wnd *sdlcanvas.Window, w *Window) {
	var eventTarget LayerTouchDelegate
	var tempEvent TouchEvent
	wnd.MouseDown = func(button, x, y int) {
		if button == 1 {
			tempEvent.X, tempEvent.Y = x, y
			tempEvent.Pressed = true
			if eventTarget = w.HitTest(tempEvent); eventTarget != nil {
				eventTarget.StartTouch(tempEvent)
			}
		}
	}
	wnd.MouseUp = func(button, x, y int) {
		if button == 1 {
			tempEvent.X, tempEvent.Y = x, y
			tempEvent.Pressed = false
			if eventTarget != nil {
				eventTarget.EndTouch(tempEvent)
				eventTarget = nil
			}
		}
	}
	wnd.MouseMove = func(x, y int) {
		if tempEvent.Pressed {
			tempEvent.X, tempEvent.Y = x, y
			if eventTarget != nil {
				eventTarget.UpdateTouch(tempEvent)
			}
		}
	}
}

func RunLoop(ctx context.Context, w *Window) error {
	// Create a macOS window via the Canvas library
	width, height := w.display.Size.X, w.display.Size.Y
	name := filepath.Base(os.Args[0])
	nativeWindow, canvas, err := sdlcanvas.CreateWindow(width, height, name)
	if err != nil {
		return err
	}
	defer nativeWindow.Close()
	bindMouseEvents(nativeWindow, w)

	scaledBuffer := image.NewRGBA(image.Rectangle{Max: w.display.Size.Mul(2)})
	nativeWindow.MainLoop(func() {
		select {
		case <-ctx.Done():
			println("Cancel")
			nativeWindow.Close()
		default:
		}
		w.update(func(region *image.RGBA) {
			var dstRect image.Rectangle
			dstRect.Min = region.Rect.Min.Mul(2)
			dstRect.Max = region.Rect.Max.Mul(2)
			draw.NearestNeighbor.Scale(scaledBuffer, dstRect, region, region.Rect, draw.Src, nil)
		})
		canvas.PutImageData(scaledBuffer, 0, 0)
	})

	return nil
}
