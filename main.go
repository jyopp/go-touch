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
}

type DisplayBuffer struct {
	Width, Height int32
	Screenbuffer  []byte
	Framebuffer   *os.File
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

func (c *TouchscreenCalibration) x(raw int32) int32 {
	return c.ScreenWidth * (raw - c.Left) / (c.Right - c.Left)
}

func (c *TouchscreenCalibration) y(raw int32) int32 {
	return c.ScreenHeight - c.ScreenHeight*(raw-c.Bottom)/(c.Top-c.Bottom)
}

func (c *TouchscreenCalibration) z(raw int32) int32 {
	return 0xFF * (raw - c.Strong) / (c.Weak - c.Strong)
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
				// VSync
				if currentEvent.X > 0 && currentEvent.Y > 0 {
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
					if e.Value > 0 {
						currentEvent.Y = calibration.y(e.Value)
					}
				case 1:
					// X-Coordinate
					if e.Value > 0 {
						currentEvent.X = calibration.x(e.Value)
					}
				case 24:
					// Digitizer pressure
					currentEvent.Pressure = calibration.z(e.Value)
				}
			}
		}
	}
}

func pixel565(r, g, b int32) (byte, byte) {
	v := (r>>3<<8|g)>>2<<5 | b
	return byte(v), byte(v >> 8)
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

func (d *DisplayBuffer) DrawBackground() {
	var r, g, b int32
	offset := 0
	for y := int32(0); y < d.Height; y++ {
		g = 0xFF * y / d.Height
		for x := int32(0); x < d.Width; x++ {
			r = 0xFF * x / d.Width
			d.Screenbuffer[offset], d.Screenbuffer[offset+1] = pixel565(r, g, b)
			offset += 2
		}
	}
}

func (d *DisplayBuffer) DrawCrosshairs(event TouchscreenEvent) {
	x := clamp(event.X, 0, d.Width-1)
	y := clamp(event.Y, 0, d.Height-1)
	offset := 2 * (d.Width*y + x)
	d.Screenbuffer[offset], d.Screenbuffer[offset+1] = 0xFF, 0xFF
}

func main() {
	{
		var display *DisplayBuffer
		if framebuffer, err := os.OpenFile("/dev/fb1", os.O_RDWR, 0); err == nil {
			display = NewDisplayBuffer(480, 320, framebuffer)
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
			ScreenWidth:  480,
			ScreenHeight: 320,
		}
		eventChannel := make(chan TouchscreenEvent, 100)

		inputFile, err := os.Open("/dev/input/event0")
		if err != nil {
			panic(err)
		}
		go emitTouchsreenEvents(inputFile, calibration, eventChannel)

		lastEvent := TouchscreenEvent{}
		display.DrawBackground()
		// Track inputs
		var events = 0
		for {
			select {
			case event := <-eventChannel:
				if event.X != lastEvent.X || event.Y != lastEvent.Y || event.Pressed != lastEvent.Pressed {
					display.DrawCrosshairs(event)
				}
				lastEvent = event
				events++
			case <-displayTimer.C:
				// Unused
			}
		}
	}
}
