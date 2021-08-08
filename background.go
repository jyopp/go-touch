package main

type Background struct {
	BasicLayer
	// Value from 0-255 controlling the brightness of the gradient
	Brightness int
}

func NewBackground(frame Rect) *Background {
	return &Background{
		BasicLayer: *NewLayer(frame, nil),
	}
}

func (background *Background) DisplayIfNeeded(ctx LayerDrawing) {
	if background.needsRedraw {
		background.DrawLayer()
	}
	background.BasicLayer.DisplayIfNeeded(ctx)
}

func (background *Background) DrawLayer() {
	buffer := background.buffer
	bright := background.Brightness
	println("Drawing Background", buffer, bright)

	var r, g, b byte
	//	g = byte((bright * 3) / 4)
	w, h := buffer.Width, buffer.Height
	for y := 0; y < h; y++ {
		b = byte(bright * y / h)
		row := buffer.pixels[2*w*y:]
		for x := 0; x < w; x++ {
			r = byte(bright * x / w)
			g = byte(bright) - r/4 - b/2
			row[x<<1], row[(x<<1)+1] = pixel565(r, g, b)
		}
	}
	background.needsRedraw = false
	background.needsDisplay = true
}
