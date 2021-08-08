package main

type Background struct {
	BasicLayer
	// Value from 0-255 controlling the brightness of the gradient
	Brightness int
}

func NewBackground(frame Rect) *Background {
	background := &Background{}
	background.Init(frame, background)
	return background
}

func (background *Background) Draw(layer Layer, ctx LayerDrawing) {
	bright := background.Brightness

	var r, g, b byte
	//	g = byte((bright * 3) / 4)
	w, h := background.w, background.h
	row := make([]byte, 2*w)

	for y := 0; y < h; y++ {
		b = byte(bright * y / h)
		for x := 0; x < w; x++ {
			r = byte(bright * x / w)
			g = byte(bright) - r/4 - b/2
			row[2*x], row[2*x+1] = pixel565(r, g, b)
		}
		ctx.DrawRow(row, background.x, background.y+y)
	}
}
