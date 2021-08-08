package main

type TouchscreenCalibration struct {
	Left, Top, Right, Bottom int
	Weak, Strong             int
	// Cached Values for faster conversions
	convW, convH, convZ int
}

func (c *TouchscreenCalibration) Prepare(display *Display) {
	c.convW = (display.Width << 16) / (c.Right - c.Left)
	c.convH = (display.Height << 16) / (c.Top - c.Bottom)
	c.convZ = (1 << 24) / (c.Weak - c.Strong)
}

func (c *TouchscreenCalibration) Adjust(ev *TouchEvent) {
	ev.X = ((ev.X - c.Left) * c.convW) >> 16
	ev.Y = ((ev.Y - c.Bottom) * c.convH) >> 16
	ev.Pressure = ((ev.Pressure - c.Strong) * c.convZ) >> 16
}
