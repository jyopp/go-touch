package main

type LayerEvents interface {
	StartTouch(TouchEvent)
	UpdateTouch(TouchEvent)
	EndTouch(TouchEvent)
}

type Layer struct {
	Rect
	Contents     []byte
	Children     []*Layer
	Owner        interface{}
	NeedsDisplay bool
}

func NewLayer(r Rect) *Layer {
	return &Layer{
		Rect:     r,
		Contents: make([]byte, 2*r.w*r.h),
	}
}

func (layer *Layer) AddChild(child *Layer) {
	layer.Children = append(layer.Children, child)
}

func (layer *Layer) HitTest(event TouchEvent) *Layer {
	for _, child := range layer.Children {
		if hit := child.HitTest(event); hit != nil {
			return hit
		}
	}
	if event.Pressed && layer.Contains(event.X, event.Y) {
		return layer
	}
	return nil
}

func (layer *Layer) StartTouch(event TouchEvent) {
	if e, ok := layer.Owner.(LayerEvents); ok {
		e.StartTouch(event)
	}
}

func (layer *Layer) UpdateTouch(event TouchEvent) {
	if e, ok := layer.Owner.(LayerEvents); ok {
		e.UpdateTouch(event)
	}
}

func (layer *Layer) EndTouch(event TouchEvent) {
	if e, ok := layer.Owner.(LayerEvents); ok {
		e.EndTouch(event)
	}
}

func (layer *Layer) FillRGB(r, g, b byte) {
	b1, b2 := pixel565(r, g, b)
	w2 := layer.w * 2
	for i := int32(0); i < w2; {
		layer.Contents[i] = b1
		i++
		layer.Contents[i] = b2
		i++
	}
	for i := int32(1); i < layer.h; i++ {
		copy(layer.Contents[i*w2:], layer.Contents[:w2])
	}
	layer.NeedsDisplay = true
}

func (layer *Layer) FillBackgroundGradient(bright int32) {
	var r, g, b byte
	//	g = byte((bright * 3) / 4)
	w, h := layer.w, layer.h
	for y := int32(0); y < h; y++ {
		b = byte(bright * y / h)
		row := layer.Contents[2*w*y:]
		for x := int32(0); x < w; x++ {
			r = byte(bright * x / w)
			g = byte(bright) - r/4 - b/2
			row[x<<1], row[(x<<1)+1] = pixel565(r, g, b)
		}
	}
	layer.NeedsDisplay = true
}

func (layer *Layer) DrawIfNeeded(display *Display) {
	if layer.NeedsDisplay {
		layer.DrawIn(display)
	} else {
		for _, child := range layer.Children {
			child.DrawIfNeeded(display)
		}
	}
}

// DrawIn Naively draws all of the buffer's pixels to the display
func (layer *Layer) DrawIn(disp *Display) {
	if layer.x > disp.Width || layer.y > disp.Height {
		return
	}

	// Determine the range endpoints for each line in the buffer, to avoid wrapping onscreen.
	// We'll copy slices from the receiver into the display's framebuffer
	var l, t, r, b int32 = 0, 0, layer.w, layer.h
	dstL := layer.x
	dstT := layer.y
	if layer.x < 0 {
		l = -layer.x
		dstL = 0
	}
	if layer.Right() > disp.Width {
		r = disp.Width - layer.x
	}
	if layer.y < 0 {
		t = -layer.y
		dstT = 0
	}
	if layer.Bottom() > disp.Height {
		b = disp.Height - layer.y
	}

	dstOffset := 2 * (dstT*disp.Width + dstL)
	srcOffsetL := 2 * (layer.w*t + l)
	srcOffsetR := 2 * (layer.w*t + r)
	srcLine := 2 * layer.w
	dstLine := 2 * disp.Width
	// Inset value used for rounded corners
	var i int32 = 0
	for srcY := t; srcY < b; srcY++ {
		// Support rounded corners via clipping
		if layer.rounded {
			i = 2 * layer.roundRectInset(srcY)
		}
		copy(disp.FrameBuffer[dstOffset+i:], layer.Contents[srcOffsetL+i:srcOffsetR-i])
		dstOffset += dstLine
		srcOffsetL += srcLine
		srcOffsetR += srcLine
	}

	// Recurse. No fancy data structures (yet)
	for _, child := range layer.Children {
		child.DrawIn(disp)
	}
	layer.NeedsDisplay = false
}
