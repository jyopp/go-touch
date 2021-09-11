package touch

import (
	"image"
	"os"
)

type Display struct {
	Size        image.Point
	FrameBuffer []byte
	DeviceFile  *os.File

	// Digitzer values for screen corners, and for weak / strong press
	Calibration *TouchscreenCalibration
}

// Clear writes zeros to the framebuffer without performing
// any drawing or buffering. This should generally not be necessary.
func (d *Display) Clear() {
	for idx := range d.FrameBuffer {
		d.FrameBuffer[idx] = 0x00
	}
}
