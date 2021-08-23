package fbui

import "context"

func RunLoop(ctx context.Context, w *Window, e *EventStream) error {
	// Draw the initial state of display
	w.update()
	// Start sending events to the event channel
	go e.inputReadLoop()

	var eventTarget LayerTouchDelegate
outer:
	for {
		select {
		case event := <-e.Events:
			w.Calibrate(&event)
			if event.Pressed {
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
		case <-w.redrawCh:
			w.update()
		case <-ctx.Done():
			break outer
		}
	}
	return nil
}
