package touch

// TODO: This needs to be based around an affine transform
type TouchscreenCalibration struct {
	MinX, MinY, MaxX, MaxY int
	Weak, Strong           int
	// Cached Values for faster conversions
	convW, convH, convZ int
	swapAxes            bool
}

// prepare updates cached values used to adjust touch events.
// Must call after any changes to Min/Max values or orientation.
func (c *TouchscreenCalibration) prepare(w, h int) {
	c.convW = (w << 16) / (c.MaxX - c.MinX)
	c.convH = (h << 16) / (c.MinY - c.MaxY)
	c.convZ = (1 << 24) / (c.Weak - c.Strong)
}

func (c *TouchscreenCalibration) Adjust(ev *TouchEvent) {
	// Nil calibrations are callable with no effect.
	if c == nil {
		return
	}

	if c.swapAxes {
		ev.X, ev.Y = ev.Y, ev.X
	}
	ev.X = ((ev.X - c.MinX) * c.convW) >> 16
	ev.Y = ((ev.Y - c.MaxY) * c.convH) >> 16
	ev.Pressure = ((ev.Pressure - c.Strong) * c.convZ) >> 16
}

func (c *TouchscreenCalibration) orient(angle int) {
	switch angle {
	case 0:
		break
	case 90:
		c.swapAxes = true
		// Reverse Y-Direction
		c.MaxY, c.MinY = c.MinY, c.MaxY
	case 270:
		c.swapAxes = true
		// Reverse X-Direction
		c.MinX, c.MaxX = c.MaxX, c.MinX
	case 180:
		// Reverse both axes
		c.MinX, c.MaxX = c.MaxX, c.MinX
		c.MaxY, c.MinY = c.MinY, c.MaxY
	default:
		panic("unsupported rotation angle")
	}
}
