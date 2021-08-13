package main

import (
	"image/color"
	"image/draw"
)

type OpaqueLayer struct {
	BasicLayer
	Background color.Color
}

func (l *OpaqueLayer) Init(frame Rect, background color.Color) *OpaqueLayer {
	l.BasicLayer.Init(frame, l)
	l.Background = background
	return l
}

func (l *OpaqueLayer) Draw(layer Layer, ctx DrawingContext) {
	op := draw.Src
	if _, _, _, a := l.Background.RGBA(); a < 0xFFFF {
		op = draw.Over
	}
	ctx.Fill(l.Rect, l.Background, op)
}
