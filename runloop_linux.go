package touch

import (
	"context"
	"os"
)

func (runloop *RunLoop) platformInit() {
	// For Linux, open the eventfile and start reading it.
	eventFile, err := os.Open("/dev/input/event0")
	if err != nil {
		panic(err)
	}

	var e EventStream
	e.Init()
	runloop.events = e.Events
	go e.inputReadLoop(eventFile)
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
