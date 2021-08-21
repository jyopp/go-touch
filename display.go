package fbui

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"sort"
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
	DirtyRects  []image.Rectangle

	// Digitzer values for screen corners, and for weak / strong press
	Calibration TouchscreenCalibration
}

// TODO: Read info from ioctl, which is nontrivial.
// See https://www.kernel.org/doc/html/latest/fb/api.html
// See https://github.com/torvalds/linux/blob/master/include/uapi/linux/fb.h
//
// const FBIOGET_VSCREENINFO = 0x4600
// const FBIOPUT_VSCREENINFO = 0x4601
// const FBIOGET_FSCREENINFO = 0x4602
// rv, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
//                               uintptr(FBIOGET_FSCREENINFO),
//                               uintptr(unsafe.Pointer(&fixedScreenInfo)))

func (d *Display) Init(w, h, rotation int, framebuffer *os.File, calibration TouchscreenCalibration) {
	// Experimental MMAP, probably not robust.
	fd := int(framebuffer.Fd())
	const protRW = syscall.PROT_WRITE | syscall.PROT_READ

	fbData, err := syscall.Mmap(fd, 0, int(2*w*h), protRW, syscall.MAP_SHARED)
	if err != nil {
		panic(fmt.Errorf("can't mmap framebuffer: %v", err))
	}

	calibration.orient(rotation)
	if calibration.swapAxes {
		// NOTE: This swaps Display buffers' dimensions too.
		w, h = h, w
	}

	bounds := image.Rectangle{Max: image.Point{w, h}}
	calibration.prepare(w, h)

	*d = Display{
		Size:        bounds.Max,
		FrameBuffer: fbData,
		DeviceFile:  framebuffer,
		Layers:      []Layer{},
		DrawBuffer:  NewDisplayBuffer(d, bounds),
		DirtyRects:  []image.Rectangle{bounds},
		Calibration: calibration,
	}
}

func (d *Display) Bounds() image.Rectangle {
	return image.Rectangle{Max: d.Size}
}

// Add a layer to the display
func (d *Display) AddLayer(layer Layer) {
	d.Layers = append(d.Layers, layer)
}

// Top-level dispatch
func (d *Display) HitTest(event TouchEvent) LayerTouchDelegate {
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

// update traverses the layer hierarchy, displaying any layers
// that need to be displayed. If any layers are displayed, a
// superset of all drawn rects is flushed to the display.
func (d *Display) update() {
	start := time.Now()
	for _, layer := range d.Layers {
		if clip := d.DrawBuffer.Clip(layer.Frame()); clip != nil {
			layer.DisplayIfNeeded(clip)
		}
	}
	drawn := time.Now()

	// Don't delegate to Flush because we're dumping diagnostic logs...
	d.mergeDirtyRects()
	for _, rect := range d.DirtyRects {
		d.flushRect(rect)
	}

	if time.Since(start).Milliseconds() > 0 {
		fmt.Printf(
			"Updated: Draw %dms / Flush %dms in %v\n",
			drawn.Sub(start).Milliseconds(),
			time.Since(drawn).Milliseconds(),
			d.DirtyRects,
		)
	}
	d.DirtyRects = d.DirtyRects[:0]
}

// TODO: This can be made much more complex, returning an array-of-rects
// For example if r1.Intersect(r2).(Min,Max).Y == r2.(Min,Max).Y, the
// intersecting part of r2 should be removed and the exclusion returned.
func shouldMergeDrawRects(r1, r2 image.Rectangle) bool {
	if !r1.Overlaps(r2) {
		return false
	} else if r1.Min.Y <= r2.Min.Y {
		// r1.Min is above or at r2.Min
		return r2.Max.Y <= r1.Min.Y
	} else {
		return r2.Max.Y >= r1.Max.Y
	}
}

func (d *Display) mergeDirtyRects() {
	if len(d.DirtyRects) == 0 {
		return
	}
	// Sort rects by their Min.Y in ascending order
	sort.Slice(d.DirtyRects, func(i, j int) bool {
		return d.DirtyRects[i].Min.Y < d.DirtyRects[j].Min.Y
	})
	// Reduce all overlapping rects.
	// This is a heuristic, vulnerable to some worst-case patterns.
	for idx, rect := range d.DirtyRects[1:] {
		if rect.Overlaps(d.DirtyRects[idx]) {
			d.DirtyRects[idx+1] = rect.Union(d.DirtyRects[idx])
			d.DirtyRects[idx] = image.Rectangle{}
		}
	}
	sort.Slice(d.DirtyRects, func(i, j int) bool {
		rI, rJ := d.DirtyRects[i], d.DirtyRects[j]
		if rI.Empty() {
			return false
		} else if rJ.Empty() {
			return true
		}
		return d.DirtyRects[i].Min.Y < d.DirtyRects[j].Min.Y
	})
	// Truncate all the empty rects out.
	i := len(d.DirtyRects)
	for i > 0 && d.DirtyRects[i-1].Empty() {
		i--
	}
	d.DirtyRects = d.DirtyRects[:i]
}

// SetDirty expands or appends a dirty rect to include all pixels in rect.
func (d *Display) SetDirty(rect image.Rectangle) {
	for idx := range d.DirtyRects {
		if shouldMergeDrawRects(rect, d.DirtyRects[idx]) {
			d.DirtyRects[idx] = rect.Union(d.DirtyRects[idx])
			return
		}
	}
	d.DirtyRects = append(d.DirtyRects, rect)
}

// Redraw erases the contents of the DrawBuffer and unconditonally
// redraws all layers.
// The entire DrawBuffer is flushed to the display before returning.
func (d *Display) Redraw() {
	buf := d.DrawBuffer
	buf.Reset(color.RGBA{})
	for _, layer := range d.Layers {
		if clip := buf.Clip(layer.Frame()); clip != nil {
			layer.Display(clip)
		}
	}
	d.flushRect(buf.Rect)
	d.DirtyRects = d.DirtyRects[:0]
}

// Flush writes downsampled pixels in DirtyRect to the Framebuffer.
// If DirtyRect is empty, this function returns immediately.
// Upon return, dirtyRect is always empty.
func (d *Display) Flush() {
	if len(d.DirtyRects) == 0 {
		return
	}
	d.mergeDirtyRects()
	for _, rect := range d.DirtyRects {
		d.flushRect(rect)
	}
	d.DirtyRects = d.DirtyRects[:0]
}

func (d *Display) flushRect(dirty image.Rectangle) {
	if dirty.Empty() {
		// Nothing to draw
		return
	}
	// println("Flushing Rect", d.DirtyRect.String())

	buf := d.DrawBuffer

	mask := CornerMask{d.Bounds(), 9}
	// If any of the corners were drawn, mask them out before flushing
	if v, h := mask.OpaqueRects(); !(dirty.In(v) || dirty.In(h)) {
		mask.EraseCorners(buf)
	}

	min, max := dirty.Min, dirty.Max
	fbStride := buf.Stride / 2

	rowL, rowR := 4*min.X, 4*max.X
	if rowR > buf.Stride {
		rowR = buf.Stride
	}

	left := min.Y * fbStride
	for y := min.Y; y < max.Y; y++ {
		fbRow := d.FrameBuffer[left : left+fbStride : left+fbStride]
		left += fbStride

		row := buf.GetRow(y)
		for i := rowL; i < rowR; i += 4 {
			sPxl := row[i : i+4 : i+4]
			// Smush the pixel down to 16 bits and assign.
			fbRow[i>>1], fbRow[i>>1+1] = pixel565(sPxl[0], sPxl[1], sPxl[2])
		}
	}
}

func pixel565(r, g, b byte) (byte, byte) {
	return ((g & 0b00011100) << 3) | b>>3, (r & 0b11111000) | g>>5
}
