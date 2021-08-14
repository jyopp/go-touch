package main

import (
	"fmt"
	"image/color"
	"image/jpeg"
	"net/http"
	"os"
	"sync"
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
			Left: 235, Right: 3750,
			Top: 3800, Bottom: 80,
			Weak: 180, Strong: 80,
		}
		calibration.Prepare(display)

		background := &Background{}
		background.Init(display.Bounds())
		background.Radius = 8
		background.Brightness = 0xEE

		downloadBackground := func(button *Button) {
			button.Label = "Downloading…"
			button.SetNeedsDisplay()
			display.Update()

			const url = "https://news-cdn.softpedia.com/images/news2/here-are-all-iphone-and-mac-wallpapers-ever-released-by-apple-528707-3.jpg"
			if resp, err := http.Get(url); err == nil {
				defer resp.Body.Close()
				if wallpaper, err := jpeg.Decode(resp.Body); err == nil {
					imageLayer := &ImageLayer{}
					imageLayer.Init(background.Bounds(), wallpaper)
					imageLayer.Background = color.RGBA{R: 0x55, G: 0x55, B: 0x55, A: 0xFF}
					background.InsertChild(imageLayer, 0)
					button.Label = "Loaded"
				} else {
					button.Label = "Decode Err"
				}
			} else {
				button.Label = "HTTP Err"
			}
			button.Disabled = true
			button.SetNeedsDisplay()
			display.Update()
		}

		buttonArea := LayoutRect{background.Rectangle.Inset(10)}
		transparentWhite := color.RGBA{R: 0x99, G: 0x99, B: 0x99, A: 0x99}

		statusArea := &BasicLayer{}
		statusArea.SetFrame(buttonArea.Slice(40, 10, fromBottom).Rectangle)
		statusArea.Background = transparentWhite
		statusArea.Radius = 5
		background.AddChild(statusArea)

		icon, _ := Resources.ReadPNG("chevron-down.png")
		for idx, rect := range buttonArea.Divide(3, 10, fromTop) {
			for idx2, rect := range rect.Divide(2, 10, fromLeft) {
				num := 2*idx + idx2
				button := &Button{}
				button.Init(rect.Rectangle)
				if num == 0 {
					button.Label = "Wallpaper"
					button.Icon, _ = Resources.ReadPNG("hex-cluster.png")
					var once sync.Once
					button.OnTap = func() {
						go once.Do(func() {
							downloadBackground(button)
						})
					}
				} else {
					button.Label = fmt.Sprintf("Button %d", 2*idx+idx2)
					button.Icon = icon
					button.OnTap = func() {
						fmt.Printf("Tapped %s\n", button.Label)
					}
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

		var eventTarget LayerTouchDelegate

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
