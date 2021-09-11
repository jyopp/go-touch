package touch

import (
	"fmt"
	"image"
	"os"
	"syscall"
)

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

func (d *Display) Init(w, h, rotation int, framebufferFile string, calibration *TouchscreenCalibration) {
	// Width and height are screen's 'natural' dimensions.
	// They will be swapped if needed based on the rotationAngle provided.

	// Open the framebuffer and get a file descriptor for it.
	framebuffer, err := os.OpenFile(framebufferFile, os.O_RDWR, 0)
	if err != nil {
		panic(err)
	}

	fd := int(framebuffer.Fd())
	const protRW = syscall.PROT_WRITE | syscall.PROT_READ

	// Experimental MMAP, probably not robust.
	fbData, err := syscall.Mmap(fd, 0, int(2*w*h), protRW, syscall.MAP_SHARED)
	if err != nil {
		panic(fmt.Errorf("can't mmap framebuffer: %v", err))
	}

	calibration.orient(rotation)
	if calibration.swapAxes {
		// NOTE: This swaps Display buffers' dimensions too.
		w, h = h, w
	}

	calibration.prepare(w, h)

	*d = Display{
		Size:        image.Point{w, h},
		FrameBuffer: fbData,
		DeviceFile:  framebuffer,
		Calibration: calibration,
	}
}

func (d *Display) Close() {
	d.Clear()
	d.DeviceFile.Close()
}

func (d *Display) render(buf *image.RGBA) {
	rect := buf.Rect
	if rect.Empty() {
		// Nothing to draw
		return
	}
	// println("Sending to framebuffer:", rect.String())

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
