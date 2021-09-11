package touch

import (
	"context"
	"os"
)

func RunLoop(ctx context.Context, w *Window) error {
	// For Linux, open the eventfile and start reading it.
	var e EventStream
	if eventFile, err := os.Open("/dev/input/event0"); err == nil {
		e.Init(eventFile)
		defer eventFile.Close()
		// events.dump = true
	} else {
		return err
	}

	// Draw the initial state of display
	w.update(w.display.render)

	// Start sending events to the event channel
	go e.inputReadLoop()

	var eventTarget LayerTouchDelegate
	var touchCanceled bool
	cancelTouch := func() {
		touchCanceled = true
		if eventTarget != nil {
			eventTarget.EndTouch(event)
			eventTarget = nil
		}
	}

outer:
	for {
		select {
		case event := <-e.Events:
			w.Calibrate(&event)
			event.Cancel = cancel
			if touchCanceled {
				// Ignore events until touch up
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
			w.update(w.display.render)
		case <-ctx.Done():
			break outer
		}
	}
	return nil
}
