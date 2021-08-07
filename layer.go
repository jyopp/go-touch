package main

type Layer struct {
	Rect
	Contents []byte
}

func NewLayer(r Rect) *Layer {
	return &Layer{
		Rect:     r,
		Contents: make([]byte, 2*r.w*r.h),
	}
}

func (buf *Layer) FillRGB(r, g, b byte) {
	b1, b2 := pixel565(r, g, b)
	w2 := buf.w * 2
	for i := int32(0); i < w2; {
		buf.Contents[i] = b1
		i++
		buf.Contents[i] = b2
		i++
	}
	for i := int32(1); i < buf.h; i++ {
		copy(buf.Contents[i*w2:], buf.Contents[:w2])
	}
}

// DrawIn Naively draws all of the buffer's pixels to the display
func (buf *Layer) DrawIn(disp *Display) {
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
		copy(disp.FrameBuffer[dstOffset+i:], buf.Contents[srcOffsetL+i:srcOffsetR-i])
		dstOffset += 2 * disp.Width
		srcOffsetL += 2 * buf.w
		srcOffsetR += 2 * buf.w
	}
}

// private function; Returns the number of pixels that should be clipped from a given line
func (r Rect) cornerInset(line int32) int32 {
	if line > r.h/2 {
		line = (r.h - 1) - line
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

func (buf *Layer) DrawBackgroundGradient(bright int32) {
	var r, g, b byte
	//	g = byte((bright * 3) / 4)
	w, h := buf.w, buf.h
	for y := int32(0); y < h; y++ {
		b = byte(bright * y / h)
		row := buf.Contents[2*w*y:]
		for x := int32(0); x < w; x++ {
			r = byte(bright * x / w)
			g = byte(bright) - r/4 - b/2
			row[x<<1], row[(x<<1)+1] = pixel565(r, g, b)
		}
	}
}
