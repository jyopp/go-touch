package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"syscall"
	"time"
)

// Eventually, perhaps Display should fully conform to LayerDrawing...

type Display struct {
	Size        image.Point
	FrameBuffer []byte
	DeviceFile  *os.File
	Layers      []Layer
	DrawBuffer  *DisplayBuffer
	DirtyRect   image.Rectangle
}

func NewDisplay(w, h int, framebuffer *os.File) *Display {
	// Experimental MMAP, probably not robust.
	data, err := syscall.Mmap(int(framebuffer.Fd()), 0, int(2*w*h), syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		panic("Can't get framebuffer")
	}
	display := &Display{
		Size:        image.Point{w, h},
		FrameBuffer: data,
		DeviceFile:  framebuffer,
		Layers:      []Layer{},
	}
	display.DrawBuffer = NewDisplayBuffer(display, display.Bounds())
	display.DirtyRect = display.DrawBuffer.Rect
	return display
}

func (d *Display) Bounds() image.Rectangle {
	return image.Rectangle{Max: d.Size}
}

// Add a layer to the display
func (d *Display) AddLayer(layer Layer) {
	d.Layers = append(d.Layers, layer)
}

// Top-level dispatch
func (d *Display) HitTest(event TouchEvent) TouchTarget {
	for _, layer := range d.Layers {
		if target := layer.HitTest(event); target != nil {
			return target
		}
	}
	return nil
}

// Clear writes zeros to the framebuffer without performing
// any drawing or buffering. This should generally not be necessary.
func (d *Display) Clear() {
	for idx := range d.FrameBuffer {
		d.FrameBuffer[idx] = 0x00
	}
}

// Update traverses the layer hierarchy, displaying any layers
// that need to be displayed. If any layers are displayed, a
// superset of all drawn rects is flushed to the display.
func (d *Display) Update() {
	start := time.Now()
	for _, layer := range d.Layers {
		if clip := d.DrawBuffer.Clip(layer.Frame()); clip != nil {
			layer.DisplayIfNeeded(clip)
		}
	}
	drawn := time.Now()
	drewRect := d.DirtyRect
	d.Flush()

	if time.Since(start).Milliseconds() > 0 {
		fmt.Printf(
			"Updated: Draw %dms / Flush %dms in %v\n",
			drawn.Sub(start).Milliseconds(),
			time.Since(drawn).Milliseconds(),
			drewRect,
		)
	}
}

// SetDirty expands the dirty rect to include all pixels in rect.
func (d *Display) SetDirty(rect image.Rectangle) {
	d.DirtyRect = d.DirtyRect.Union(rect)
}

// Redraw erases the contents of the DrawBuffer and unconditonally
// redraws all layers.
// The entire DrawBuffer is flushed to the display before returning.
func (d *Display) Redraw() {
	buf := d.DrawBuffer
	buf.Reset(color.Transparent)
	d.DirtyRect = buf.Rect
	for _, layer := range d.Layers {
		if clip := buf.Clip(layer.Frame()); clip != nil {
			layer.Display(clip)
		}
	}
	d.Flush()
}

// Flush wirtes downsampled pixels in DirtyRect to the Framebuffer.
// If DirtyRect is empty, this function returns immediately.
// Upon return, dirtyRect is always empty.
func (d *Display) Flush() {
	if d.DirtyRect.Empty() {
		// Nothing to draw
		return
	}
	// println("Flushing Rect", d.DirtyRect.String())

	min, max := d.DirtyRect.Min, d.DirtyRect.Max
	buf := d.DrawBuffer

	fbStride := d.DrawBuffer.Stride / 2

	rowL, rowR := 4*min.X, 4*max.X
	if rowR > d.DrawBuffer.Stride {
		rowR = d.DrawBuffer.Stride
	}

	for y := min.Y; y < max.Y; y++ {
		left := y * fbStride
		fbRow := d.FrameBuffer[left : left+fbStride : left+fbStride]

		row := buf.GetRow(y)
		for i := rowL; i < rowR; i += 4 {
			sPxl := row[i : i+4 : i+4]
			// Smush the pixel down to 16 bits and assign.
			fbRow[i>>1], fbRow[i>>1+1] = pixel565(sPxl[0], sPxl[1], sPxl[2])
		}
	}

	d.DirtyRect = image.Rectangle{}
}

func pixel565(r, g, b byte) (byte, byte) {
	return ((g << 3) & 0b11100000) | b>>3, (r & 0b11111000) | (g >> 5)
}
