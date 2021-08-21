package fbui

import (
	"fmt"
	"image"
	"os"
	"syscall"
)

// Eventually, perhaps Display should fully conform to LayerDrawing...

type Display struct {
	Size        image.Point
	FrameBuffer []byte
	DeviceFile  *os.File

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
		Calibration: calibration,
	}
}

func (d *Display) Bounds() image.Rectangle {
	return image.Rectangle{Max: d.Size}
}

// Clear writes zeros to the framebuffer without performing
// any drawing or buffering. This should generally not be necessary.
func (d *Display) Clear() {
	for idx := range d.FrameBuffer {
		d.FrameBuffer[idx] = 0x00
	}
}

func (d *Display) render(buf *image.RGBA) {
	rect := buf.Rect
	if rect.Empty() {
		// Nothing to draw
		return
	}
	// println("Flushing Rect", d.DirtyRect.String())

	min, max := rect.Min, rect.Max
	fbStride := buf.Stride / 2

	left := min.Y * fbStride
	for y := min.Y; y < max.Y; y++ {
		fbRow := d.FrameBuffer[left+min.X*2 : left+fbStride : left+fbStride]
		left += fbStride

		row := buf.Pix[buf.PixOffset(min.X, y):buf.PixOffset(max.X, y)]

		for i := 0; i < len(row); i += 4 {
			sPxl := row[i : i+4 : i+4]
			// Smush the pixel down to 16 bits and assign.
			fbRow[i>>1], fbRow[i>>1+1] = pixel565(sPxl[0], sPxl[1], sPxl[2])
		}
	}
}

func pixel565(r, g, b byte) (byte, byte) {
	return ((g & 0b00011100) << 3) | b>>3, (r & 0b11111000) | g>>5
}
