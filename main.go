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
		background.Init(display.Bounds(), 0xEE)

		buttonArea := LayoutRect{background.Rectangle.Inset(10)}
		// transparentWhite := color.RGBA{R: 0x99, G: 0x99, B: 0x99, A: 0x99}

		statusArea := &BasicLayer{}
		statusArea.SetFrame(buttonArea.Slice(40, 10, fromBottom).Rectangle)
		statusArea.Background = color.White
		statusArea.Radius = 5
		background.AddChild(statusArea)

		statusText := &TextLayer{}
		statusText.Init(statusArea.Rectangle.Inset(8), systemBoldFont, 11.0)
		statusText.Text = "Status Text Test"
		statusArea.AddChild(statusText)

		setStatusText := func(text string) {
			statusText.Text = text
			statusArea.SetNeedsDisplay()
		}

		downloadBackground := func(button *Button) {
			button.SetState(stateDisabled)
			button.SetNeedsDisplay()
			setStatusText("Downloading Wallpaper Image…")
			display.Update()

			const url = "https://news-cdn.softpedia.com/images/news2/here-are-all-iphone-and-mac-wallpapers-ever-released-by-apple-528707-3.jpg"
			if resp, err := http.Get(url); err == nil {
				defer resp.Body.Close()
				if wallpaper, err := jpeg.Decode(resp.Body); err == nil {
					imageLayer := &ImageLayer{}
					imageLayer.Init(background.Bounds(), wallpaper)
					imageLayer.Background = color.RGBA{R: 0x55, G: 0x55, B: 0x55, A: 0xFF}
					background.InsertChild(imageLayer, 0)
					setStatusText("Loaded Wallpaper")
				} else {
					setStatusText("Decode Error: " + err.Error())
				}
			} else {
				setStatusText("HTTP Error: " + err.Error())
			}
			display.Update()
		}

		icon, _ := Resources.ReadPNG("chevron-down.png")
		for idx, rect := range buttonArea.Divide(3, 10, fromTop) {
			for idx2, rect := range rect.Divide(2, 10, fromLeft) {
				num := 2*idx + idx2
				button := &Button{}
				button.Init(rect.Rectangle)
				if num == 0 {
					button.Label.Text = "Wallpaper"
					button.Icon, _ = Resources.ReadPNG("hex-cluster.png")
					var once sync.Once
					button.OnTap = func() {
						once.Do(func() {
							go downloadBackground(button)
						})
					}
				} else {
					button.Label.Text = fmt.Sprintf("Button %d", 2*idx+idx2)
					button.Icon = icon
					button.OnTap = func() {
						statusText.Text = fmt.Sprintf("Tapped %s", button.Label.Text)
						statusArea.SetNeedsDisplay()
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
