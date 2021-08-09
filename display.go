package main

import (
	"os"
	"syscall"
)

// Eventually, perhaps Display should fully conform to LayerDrawing...

type Display struct {
	Width, Height int
	FrameBuffer   []byte
	DeviceFile    *os.File
	Layers        []Layer
}

func NewDisplay(w, h int, framebuffer *os.File) *Display {
	// Experimental MMAP, probably not robust.
	data, err := syscall.Mmap(int(framebuffer.Fd()), 0, int(2*w*h), syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		panic("Can't get framebuffer")
	}
	return &Display{
		Width:       w,
		Height:      h,
		FrameBuffer: data,
		DeviceFile:  framebuffer,
		Layers:      []Layer{},
	}
}

func (d *Display) Bounds() Rect {
	return Rect{x: 0, y: 0, w: d.Width, h: d.Height}
}

func (d *Display) DrawPixel(x, y int, r, g, b byte) {
	if x < 0 || y < 0 || x >= d.Width || y >= d.Height {
		return
	}
	var pixel [2]byte
	pixel[0], pixel[1] = pixel565(r, g, b)
	copy(d.FrameBuffer[2*(d.Width*y+x):], pixel[:])
}

func (d *Display) AddLayer(layer Layer) {
	d.Layers = append(d.Layers, layer)
}

func (d *Display) HitTest(event TouchEvent) TouchTarget {
	for _, layer := range d.Layers {
		if target := layer.HitTest(event); target != nil {
			return target
		}
	}
	return nil
}

func (d *Display) Clear() {
	for idx := range d.FrameBuffer {
		d.FrameBuffer[idx] = 0x00
	}
}

func (d *Display) Update() {
	for _, layer := range d.Layers {
		layer.DisplayIfNeeded(d.Subrect(layer.Frame()))
	}
}

func (d *Display) Subrect(r Rect) *DisplayRect {
	return &DisplayRect{
		Rect:    r,
		display: d,
	}
}
