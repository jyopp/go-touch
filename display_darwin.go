package touch

import "image"

func (d *Display) Init(w, h, rotation int, framebufferFile string, calibration *TouchscreenCalibration) {
	calibration.orient(rotation)
	if calibration.swapAxes {
		// NOTE: This swaps Display buffers' dimensions too.
		w, h = h, w
	}

	calibration.prepare(w, h)

	// Return a mostly-empty display record on Mac, as many things
	// are handled via native bindings, or by the window manager.
	*d = Display{
		Size:        image.Point{w, h},
		Calibration: nil,
	}
}

func (d *Display) Close() {
}
