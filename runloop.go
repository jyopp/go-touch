package touch

import "context"

var (
	MainRunLoop RunLoop
)

type RunLoop struct {
	Window *Window
	Tasks  chan<- func()
	tasks  chan func()
	events chan TouchEvent
}

func (runloop *RunLoop) Init(window *Window) {
	if runloop != &MainRunLoop {
		panic("Only the main RunLoop may be Initialized")
	}
	runloop.tasks = make(chan func(), 100)
	runloop.Tasks = runloop.tasks
	runloop.Window = window
	runloop.platformInit()
}

func (runloop *RunLoop) runInner(ctx context.Context) {
	var eventTarget LayerTouchDelegate
	var touchCanceled bool
	cancelTouch := func() {
		touchCanceled = true
		if eventTarget != nil {
			eventTarget.CancelTouch()
			eventTarget = nil
		}
	}

	win := runloop.Window
	runloop.updateDisplay()

outer:
	for {
		select {
		case event := <-runloop.events:
			win.Calibrate(&event)
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
					if eventTarget = win.HitTest(event); eventTarget != nil {
						eventTarget.StartTouch(event)
					}
				}
			} else {
				if eventTarget != nil {
					eventTarget.EndTouch(event)
					eventTarget = nil
				}
			}
		case task := <-runloop.tasks:
			task()
		case <-win.redrawCh:
			runloop.updateDisplay()
		case <-ctx.Done():
			runloop.cleanup()
			break outer
		}
	}
}
