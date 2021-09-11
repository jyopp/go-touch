package touch

import (
	"context"
	"os"
)

func (runloop *RunLoop) platformInit() {
	// For Linux, open the eventfile and start reading it.
	var e EventStream
	if eventFile, err := os.Open("/dev/input/event0"); err == nil {
		e.Init(eventFile)
	} else {
		panic(err)
	}

	runloop.events = e.Events
	go e.inputReadLoop()
}

func (runloop *RunLoop) updateDisplay() {
	win := runloop.Window
	win.update(win.display.render)
}

func (runloop *RunLoop) cleanup() {
	// No additional cleanup on Linux
}

func (runloop *RunLoop) Run(ctx context.Context) {
	runloop.runInner(ctx)
}
