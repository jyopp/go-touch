package main

import (
	"fmt"
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

		calibration := TouchscreenCalibration{
			Left: 235, Right: 3750,
			Top: 3800, Bottom: 80,
			Weak: 180, Strong: 80,
		}
		calibration.Prepare(display)

		// start := time.Now()
		// for i := int32(0); i <= 0xFF; i += 8 {
		// 	display.DrawBackground(i)
		// }
		// fmt.Printf("Screen draw averaged %0.2dms over 32 passes\n", time.Since(start).Milliseconds()/32.0)

		buttonArea := display.Background.Inset(10, 10)
		for idx, rect := range buttonArea.GridVCount(6, 10) {
			button := NewButton(rect)
			button.Label = fmt.Sprintf("Button %d", idx)
			button.DrawLayer()
			button.OnTap = func() {
				fmt.Printf("Tapped %s\n", button.Label)
			}
			display.Background.AddChild(button.Layer)
		}

		display.Background.FillBackgroundGradient(0xE0)
		display.Redraw()

		// Track inputs
		inputFile, err := os.Open("/dev/input/event0")
		if err != nil {
			panic(err)
		}

		events := NewEventStream(inputFile)
		// events.dump = true
		go events.EventLoop()

		var eventTarget *Layer

		for event := range events.TouchEvents {
			calibration.Adjust(&event)
			if event.Pressed {
				if eventTarget != nil {
					eventTarget.UpdateTouch(event)
				} else {
					eventTarget = display.Background.HitTest(event)
					if eventTarget != nil {
						eventTarget.StartTouch(event)
					}
				}
			} else {
				if eventTarget != nil {
					eventTarget.EndTouch(event)
					eventTarget = nil
				}
			}
			display.Draw()
		}
	}
}
