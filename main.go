package main

import (
	"encoding/binary"
	"io"
	"os"
	"syscall"
	"time"
)

type InputEvent struct {
	Time  syscall.Timeval
	Type  uint16
	Code  uint16
	Value int32
}

type TouchscreenEvent struct {
	Pressed  bool
	X, Y     int32
	Pressure int32
}

type TouchscreenCalibration struct {
	// Digitzer values for screen corners, and for weak / strong press
	Left, Top, Right, Bottom  int32
	Weak, Strong              int32
	ScreenWidth, ScreenHeight int32
	// Cached Values for faster conversions
	convW, convH, convZ int32
}

type DisplayBuffer struct {
	Width, Height int32
	Screenbuffer  []byte
	Framebuffer   *os.File
}

type Rect struct {
	x, y    int32
	w, h    int32
	rounded bool
}

func (r Rect) Bottom() int32 {
	return r.y + r.h
}

func (r Rect) Right() int32 {
	return r.x + r.w
}

func (r Rect) Contains(x, y int32) bool {
	return x >= r.x && y >= r.y && x < r.x+r.w && y < r.y+r.h
}

type RectBuffer struct {
	Rect
	Buffer []byte
}

type Button struct {
	*RectBuffer
	Pressed bool
}

func NewBuffer(r Rect) *RectBuffer {
	return &RectBuffer{
		Rect:   r,
		Buffer: make([]byte, 2*r.w*r.h),
	}
}

func NewButton(r Rect) *Button {
	button := &Button{
		RectBuffer: NewBuffer(r),
	}
	button.rounded = true
	return button
}

func (buf *RectBuffer) FillRGB(r, g, b byte) {
	b1, b2 := pixel565(r, g, b)
	w2 := buf.w * 2
	for i := int32(0); i < w2; {
		buf.Buffer[i] = b1
		i++
		buf.Buffer[i] = b2
		i++
	}
	for i := int32(1); i < buf.h; i++ {
		copy(buf.Buffer[i*w2:], buf.Buffer[:w2])
	}
}

// DrawIn Naively draws all of the buffer's pixels to the display
func (buf *RectBuffer) DrawIn(disp *DisplayBuffer) {
	if buf.x > disp.Width || buf.y > disp.Height {
		return
	}

	// Determine the range endpoints for each line in the buffer, to avoid wrapping onscreen.
	// We'll slice the display's Screenbuffer
	var l, t, r, b int32 = 0, 0, buf.w, buf.h
	dstL := buf.x
	dstT := buf.y
	if buf.x < 0 {
		l = -buf.x
		dstL = 0
	}
	if buf.Right() > disp.Width {
		r = disp.Width - buf.x
	}
	if buf.y < 0 {
		t = -buf.y
		dstT = 0
	}
	if buf.Bottom() > disp.Height {
		b = disp.Height - buf.y
	}
	dstOffset := 2 * (dstT*disp.Width + dstL)
	srcOffsetL := 2 * (buf.w*t + l)
	srcOffsetR := 2 * (buf.w*t + r)
	// Inset value used for rounded corners
	var i int32 = 0
	for srcY := t; srcY < b; srcY++ {
		// Support rounded corners via clipping
		if buf.rounded {
			i = 2 * buf.cornerInset(srcY)
		}
		copy(disp.Screenbuffer[dstOffset+i:], buf.Buffer[srcOffsetL+i:srcOffsetR-i])
		dstOffset += 2 * disp.Width
		srcOffsetL += 2 * buf.w
		srcOffsetR += 2 * buf.w
	}
}

// private function; Returns the number of pixels that should be clipped from a given line
func (r Rect) cornerInset(line int32) int32 {
	if line > r.h/2 {
		line = r.h - line
	}
	switch line {
	case 0:
		return 5
	case 1:
		return 3
	case 2:
		return 2
	case 3:
		return 1
	case 4:
		return 1
	}
	return 0
}

func NewDisplayBuffer(w, h int32, framebuffer *os.File) *DisplayBuffer {
	// Experimental MMAP, probably not robust.
	data, err := syscall.Mmap(int(framebuffer.Fd()), 0, int(2*w*h), syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		panic("Can't get framebuffer")
	}
	return &DisplayBuffer{
		Width:        w,
		Height:       h,
		Screenbuffer: data,
		Framebuffer:  framebuffer,
	}
}

func (c *TouchscreenCalibration) Init() {
	c.convW = (c.ScreenWidth << 16) / (c.Right - c.Left)
	c.convH = (c.ScreenHeight << 16) / (c.Top - c.Bottom)
	c.convZ = (1 << 24) / (c.Weak - c.Strong)
}

func (c *TouchscreenCalibration) Adjust(ev *TouchscreenEvent) {
	ev.X = ((ev.X - c.Left) * c.convW) >> 16
	ev.Y = ((ev.Y - c.Bottom) * c.convH) >> 16
	ev.Pressure = ((ev.Pressure - c.Strong) * c.convZ) >> 16
}

func emitTouchsreenEvents(src io.Reader, calibration TouchscreenCalibration, output chan TouchscreenEvent) {
	var currentEvent TouchscreenEvent
	rawEvents := make([]InputEvent, 2)

	for {
		rawEvents = rawEvents[:]
		if err := binary.Read(src, binary.LittleEndian, &rawEvents); err != nil {
			return
		}
		// println(len(rawEvents), "events")

		for _, e := range rawEvents {
			if e.Time.Sec == 0 {
				continue
			}
			switch e.Type {
			case 0:
				if currentEvent.X > 0 && currentEvent.Y > 0 {
					calibration.Adjust(&currentEvent)
					output <- currentEvent
				}
				currentEvent = TouchscreenEvent{}
			case 1:
				// Button event
				currentEvent.Pressed = e.Value > 0
			case 3:
				// State event
				switch e.Code {
				case 0:
					// Y-coordinate
					currentEvent.X = e.Value
				case 1:
					// X-Coordinate
					currentEvent.Y = e.Value
				case 24:
					// Digitizer pressure
					currentEvent.Pressure = e.Value
				}
			}
		}
	}
}

func pixel565(r, g, b byte) (byte, byte) {
	return ((g << 3) & 0b11100000) | b>>3, (r & 0b11111000) | (g >> 5)
}

func clamp(num, min, max int32) int32 {
	if num < min {
		return min
	}
	if num > max {
		return max
	}
	return num
}

func (d *DisplayBuffer) DrawBackground(scale int32) {
	var r, g, b byte
	w, h := d.Width, d.Height
	buf := make([]byte, 2*w*h)
	// row := make([]byte, 2*w)
	for y := int32(0); y < h; y++ {
		b = byte(scale * y / h)
		// row := d.Screenbuffer[2*w*y:]
		row := buf[2*w*y:]
		for x := int32(0); x < w; x++ {
			r = byte(scale * x / w)
			row[x<<1], row[(x<<1)+1] = pixel565(r, g, b)
		}
		// copy(d.Screenbuffer[2*w*y:], row)
	}
	copy(d.Screenbuffer, buf)
}

func (d *DisplayBuffer) DrawPixel(x, y int32, r, g, b byte) {
	x = clamp(x, 0, d.Width-1)
	y = clamp(y, 0, d.Height-1)
	var pixel [2]byte
	pixel[0], pixel[1] = pixel565(r, g, b)
	copy(d.Screenbuffer[2*(d.Width*y+x):], pixel[:])
}

func main() {
	{
		var display *DisplayBuffer
		if framebuffer, err := os.OpenFile("/dev/fb1", os.O_RDWR, 0); err == nil {
			display = NewDisplayBuffer(320, 480, framebuffer)
			defer display.Framebuffer.Close()
		} else {
			panic(err)
		}

		const fps = 10
		displayTimer := time.NewTicker(time.Second / fps)
		defer displayTimer.Stop()

		calibration := TouchscreenCalibration{
			Left: 235, Right: 3750,
			Top: 3800, Bottom: 80,
			Weak: 180, Strong: 80,
			ScreenWidth:  display.Width,
			ScreenHeight: display.Height,
		}
		calibration.Init()

		eventChannel := make(chan TouchscreenEvent, 100)

		inputFile, err := os.Open("/dev/input/event0")
		if err != nil {
			panic(err)
		}
		go emitTouchsreenEvents(inputFile, calibration, eventChannel)

		// start := time.Now()
		// for i := int32(0); i <= 0xFF; i += 8 {
		// 	display.DrawBackground(i)
		// }
		// fmt.Printf("Screen draw averaged %0.2dms over 32 passes\n", time.Since(start).Milliseconds()/32.0)

		display.DrawBackground(0xFF)

		buttons := make([]*Button, 6)
		for idx := range buttons {
			buttons[idx] = NewButton(Rect{x: 10, y: 10 + 60*int32(idx), w: display.Width - 20, h: 50})
			buttons[idx].FillRGB(0xFF, 0xFE, 0xFC)
			buttons[idx].DrawIn(display)
		}

		// Track inputs
		lastEvent := TouchscreenEvent{}
		var events = 0
		for {
			select {
			case event := <-eventChannel:
				if event.X != lastEvent.X || event.Y != lastEvent.Y || event.Pressed != lastEvent.Pressed {
					for _, button := range buttons {
						hit := button.Contains(event.X, event.Y)
						if hit != button.Pressed {
							button.Pressed = hit
							if hit {
								button.FillRGB(0x55, 0x88, 0x99)
							} else {
								button.FillRGB(0xFF, 0xFE, 0xFC)
							}
						}
						button.DrawIn(display)
					}
					display.DrawPixel(event.X, event.Y, 0, 0, 0)
				}
				lastEvent = event
				events++
			case <-displayTimer.C:
				// Unused
			}
		}
	}
}
