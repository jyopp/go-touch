package main

import (
	"image/color"
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
	model565.Fill(ctx, l.Bounds(), l.Background)
}
