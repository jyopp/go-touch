package fbui

// TODO: This needs to be based around an affine transform
type TouchscreenCalibration struct {
	Left, Top, Right, Bottom int
	Weak, Strong             int
	// Cached Values for faster conversions
	convW, convH, convZ int
}

func (c *TouchscreenCalibration) prepare(d *Display) {
	c.convW = (d.Size.X << 16) / (c.Right - c.Left)
	c.convH = (d.Size.Y << 16) / (c.Top - c.Bottom)
	c.convZ = (1 << 24) / (c.Weak - c.Strong)
}

func (c *TouchscreenCalibration) Adjust(ev *TouchEvent) {
	ev.X = ((ev.X - c.Left) * c.convW) >> 16
	ev.Y = ((ev.Y - c.Bottom) * c.convH) >> 16
	ev.Pressure = ((ev.Pressure - c.Strong) * c.convZ) >> 16
}
