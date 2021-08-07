package main

type Button struct {
	*Layer
	Pressed bool
}

func NewButton(r Rect) *Button {
	button := &Button{
		Layer: NewLayer(r),
	}
	button.rounded = true
	return button
}
