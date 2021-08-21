package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"os"
	"os/signal"
	"sync"

	ui "github.com/jyopp/fbui"
)

const (
	DefaultFont     = "Raleway-Medium.ttf"
	DefaultBoldFont = "Raleway-SemiBold.ttf"
)

func init() {
	Resources.RegisterFont(DefaultFont)
	Resources.RegisterFont(DefaultBoldFont)
}

var (
	window     = &ui.Window{}
	events     = &ui.EventStream{}
	background = &Background{}
	statusArea = &ui.BasicLayer{}
	statusText = &ui.TextLayer{}
)

func downloadBackground(button *ui.Button) {
	button.SetDisabled(true)
	button.SetNeedsDisplay()
	statusText.SetText("Downloading Wallpaper Image…")
	events.RequestDisplayUpdate()

	const url = "https://news-cdn.softpedia.com/images/news2/here-are-all-iphone-and-mac-wallpapers-ever-released-by-apple-528707-3.jpg"
	if resp, err := http.Get(url); err == nil {
		defer resp.Body.Close()
		if wallpaper, err := jpeg.Decode(resp.Body); err == nil {
			imageLayer := &ui.ImageLayer{}
			imageLayer.Init(background.Bounds(), wallpaper)
			background.InsertChild(imageLayer, 0)
			statusText.SetText("Loaded Wallpaper")
		} else {
			statusText.SetText("Decode Error: " + err.Error())
		}
	} else {
		statusText.SetText("HTTP Error: " + err.Error())
	}
	events.RequestDisplayUpdate()
}

func buildUI() {
	background.Init(window.Bounds(), 0xEE)

	buttonArea := ui.Layout(background.Rectangle).InsetBy(10, 10)
	// transparentWhite := color.RGBA{R: 0x99, G: 0x99, B: 0x99, A: 0x99}

	statusArea.SetFrame(buttonArea.Slice(40, 10, ui.FromBottom).Rectangle)
	statusArea.Background = color.White
	statusArea.Radius = 5
	background.AddChild(statusArea)

	statusText.Init(statusArea.Rectangle, DefaultBoldFont, 11.0)
	statusText.Text = "Status Text Test"
	statusText.Color = color.Gray{0x33}
	statusText.Padding = image.Point{X: 10, Y: 5}
	statusArea.AddChild(statusText)

	icon, _ := Resources.ReadPNG("chevron-down.png")
	for idx, rect := range buttonArea.Divide(2, 10, ui.FromTop) {
		for idx2, rect := range rect.Divide(3, 10, ui.FromLeft) {
			num := 3*idx + idx2
			button := &ui.Button{}
			button.Init(rect.Rectangle, DefaultFont, 15.0)
			if num == 0 {
				button.Label.Text = "Wallpaper"
				button.Icon.Image, _ = Resources.ReadPNG("hex-cluster.png")
				var once sync.Once
				button.Actions[ui.ControlTapped] = func(button *ui.Button) {
					once.Do(func() {
						go downloadBackground(button)
					})
				}
			} else {
				button.Label.Text = fmt.Sprintf("Button %d", 2*idx+idx2)
				button.Icon.Image = icon
				button.Actions[ui.ControlTapped] = func(button *ui.Button) {
					statusText.Text = fmt.Sprintf("Tapped %s", button.Label.Text)
					statusArea.SetNeedsDisplay()
				}
			}
			background.AddChild(button)
		}
	}

	window.AddChild(background)
}

var (
	// Calibration describes the behavior of the touchscreen in its natural orientation.
	// Display will swap Min & Max values as needed to match the display's rotation.
	touchCalibration = ui.TouchscreenCalibration{
		MinX: 235, MaxX: 3750,
		MinY: 3800, MaxY: 80, // Y-Axis events are bottom-to-top in the natural orientation
		Weak: 180, Strong: 80,
	}
)

func main() {
	rotationAngle := flag.Int("rotation", 0, "Rotation of the display")
	flag.Parse()

	if framebuffer, err := os.OpenFile("/dev/fb1", os.O_RDWR, 0); err == nil {
		// Width and height are screen's 'natural' dimensions.
		// They will be swapped if needed based on the rotationAngle provided.
		display := &ui.Display{}
		display.Init(320, 480, *rotationAngle, framebuffer, touchCalibration)
		defer framebuffer.Close()
		defer display.Clear()

		window.Init(display)
	} else {
		panic(err)
	}

	if eventFile, err := os.Open("/dev/input/event0"); err == nil {
		events.Init(eventFile)
		defer eventFile.Close()
		// events.dump = true
	} else {
		panic(err)
	}

	buildUI()

	signalCtx, signalCleanup := signal.NotifyContext(context.Background(), os.Interrupt)
	defer signalCleanup()

	events.DispatchLoop(window, signalCtx)
}
