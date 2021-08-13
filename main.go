package main

import (
	"fmt"
	"image/color"
	"os"
)

func main() {
	{
		var display *Display
		if framebuffer, err := os.OpenFile("/dev/fb1", os.O_RDWR, 0); err == nil {
			display = NewDisplay(320, 480, framebuffer)
			defer framebuffer.Close()
		} else {
			panic(err)
		}

		// TODO: This needs to be an affine transform
		calibration := TouchscreenCalibration{
			Left: 3750, Right: 235,
			Top: 80, Bottom: 3800,
			Weak: 180, Strong: 80,
		}
		calibration.Prepare(display)

		background := &Background{}
		background.Init(display.Bounds())
		background.radius = 8
		background.Brightness = 0xEE


		buttonArea := background.Inset(10, 10)

		statusArea := &OpaqueLayer{}
		transparentWhite := color.RGBA{R: 0x99, G: 0x99, B: 0x99, A: 0x99}
		statusArea.Init(buttonArea.SliceV(-40, 10), transparentWhite)
		statusArea.radius = 5
		background.AddChild(statusArea)

		icon, _ := Resources.ReadPNG("note.png")
		for idx, rect := range buttonArea.GridVCount(3, 10) {
			for idx2, rect := range rect.GridHCount(2, 10) {
				button := NewButton(rect)
				button.Label = fmt.Sprintf("Button %d", 2*idx+idx2)
				button.Icon = icon
				button.OnTap = func() {
					fmt.Printf("Tapped %s\n", button.Label)
				}
				background.AddChild(button)
			}
		}

		display.AddLayer(background)
		display.Update()

		// Track inputs
		inputFile, err := os.Open("/dev/input/event0")
		if err != nil {
			panic(err)
		}

		events := NewEventStream(inputFile)
		// events.dump = true
		go events.EventLoop()

		var eventTarget TouchTarget

		for event := range events.TouchEvents {
			calibration.Adjust(&event)
			if event.Pressed {
				if eventTarget != nil {
					eventTarget.UpdateTouch(event)
				} else {
					// Only when there is no current event target, hit test for one.
					if eventTarget = display.HitTest(event); eventTarget != nil {
						eventTarget.StartTouch(event)
					}
				}
			} else {
				if eventTarget != nil {
					eventTarget.EndTouch(event)
					eventTarget = nil
				}
			}
			display.Update()
		}
	}
}
