package main

import (
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

		display.Background.DrawBackgroundGradient(0xE0)
		display.Redraw()

		buttons := make([]*Button, 6)
		for idx := range buttons {
			buttons[idx] = NewButton(Rect{x: 10, y: 10 + 60*int32(idx), w: display.Width - 20, h: 50})
			buttons[idx].FillRGB(0xFF, 0xFE, 0xFC)
			buttons[idx].DrawIn(display)
		}

		// Track inputs
		inputFile, err := os.Open("/dev/input/event0")
		if err != nil {
			panic(err)
		}

		events := NewEventStream(inputFile)
		// events.dump = true
		go events.EventLoop()

		lastEvent := TouchEvent{}
		for event := range events.TouchEvents {
			calibration.Adjust(&event)
			if event.X != lastEvent.X || event.Y != lastEvent.Y || event.Pressed != lastEvent.Pressed {
				for _, button := range buttons {
					hit := event.Pressed && button.Contains(event.X, event.Y)
					if hit != button.Pressed {
						button.Pressed = hit
						if hit {
							button.FillRGB(0x55, 0xAA, 0xCC)
						} else {
							button.FillRGB(0xFF, 0xFE, 0xFC)
						}
					}
					button.DrawIn(display)
				}
				// display.DrawPixel(event.X, event.Y, 0, 0, 0)
			}
			lastEvent = event
		}
	}
}
