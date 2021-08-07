package main

import (
	"os"
	"syscall"
)

type Display struct {
	Width, Height int32
	Background    *Layer
	FrameBuffer   []byte
	DeviceFile    *os.File
}

func NewDisplay(w, h int32, framebuffer *os.File) *Display {
	// Experimental MMAP, probably not robust.
	data, err := syscall.Mmap(int(framebuffer.Fd()), 0, int(2*w*h), syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		panic("Can't get framebuffer")
	}
	return &Display{
		Width:       w,
		Height:      h,
		Background:  NewLayer(Rect{x: 0, y: 0, w: w, h: h}),
		FrameBuffer: data,
		DeviceFile:  framebuffer,
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

func (d *Display) Redraw() {
	d.Background.DrawIn(d)
}

func (d *Display) DrawPixel(x, y int32, r, g, b byte) {
	x = clamp(x, 0, d.Width-1)
	y = clamp(y, 0, d.Height-1)
	var pixel [2]byte
	pixel[0], pixel[1] = pixel565(r, g, b)
	copy(d.FrameBuffer[2*(d.Width*y+x):], pixel[:])
}
