package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
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

func showSimpleAlert(message, buttontext string, done func()) {

	layout := ui.Layout(ui.Layout(window.Rect).Aligned(image.Point{240, 180}, ui.GravityCenter))
	wrapper := &ui.BasicLayer{}
	wrapper.Background = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x22}
	wrapper.Radius = 8
	wrapper.SetFrame(layout.Rectangle.Inset(-1))

	alert := &ui.BasicLayer{}
	alert.Background = color.White
	alert.Radius = 7
	alert.SetFrame(layout.Rectangle)

	layout = layout.InsetBy(10, 10)

	button := &ui.Button{}
	button.Init(layout.Slice(48, 10, ui.FromBottom).Rectangle, DefaultBoldFont, 14.0)
	button.Colors.Normal.Background = color.RGBA{R: 0xE0, G: 0xE0, B: 0xFF, A: 0xFF}
	button.Colors.Normal.Text = color.Gray{0x44}
	button.Label.SetText(buttontext)
	button.Actions[ui.ControlTapped] = func(button *ui.Button) {
		wrapper.RemoveFromParent()
		done()
	}

	textLayer := &ui.TextLayer{}
	textLayer.Init(layout.Rectangle, DefaultFont, 15.0)
	textLayer.Gravity = ui.GravityCenter
	textLayer.SetText(message)

	alert.AddChild(textLayer)
	alert.AddChild(button)
	wrapper.AddChild(alert)
	window.AddChild(wrapper)
}

func downloadBackground(button *ui.Button) {
	button.SetDisabled(true)
	button.Invalidate()
	statusText.SetText("Downloading Wallpaper Image…")

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
}

func buildUI() {
	background.Init(window.Bounds(), 0xEE)

	buttonArea := ui.Layout(background.Rectangle).InsetBy(10, 10)
	transparentWhite := color.RGBA{R: 0xBB, G: 0xBB, B: 0xBB, A: 0xBB}

	statusArea.SetFrame(buttonArea.Slice(40, 10, ui.FromBottom).Rectangle)
	statusArea.Background = transparentWhite
	statusArea.Radius = 5
	background.AddChild(statusArea)

	statusTextRect := ui.Layout(statusArea.Rectangle).InsetBy(10, 5).Rectangle
	statusText.Init(statusTextRect, DefaultBoldFont, 11.0)
	statusText.Text = "Status Text Test"
	statusText.Color = color.Gray{0x33}
	statusArea.AddChild(statusText)

	icon, _ := Resources.ReadPNGTemplate("chevron-down.png")
	for idx, rect := range buttonArea.Divide(2, 10, ui.FromTop) {
		for idx2, rect := range rect.Divide(3, 10, ui.FromLeft) {
			num := 3*idx + idx2
			button := &ui.Button{}
			button.Init(rect.Rectangle, DefaultFont, 15.0)
			if num == 0 {
				button.Label.Text = "Wallpaper"
				button.Icon.Image, _ = Resources.ReadPNGTemplate("hex-cluster.png")
				var once sync.Once
				button.Actions[ui.ControlTapped] = func(button *ui.Button) {
					once.Do(func() {
						go downloadBackground(button)
					})
				}
			} else {
				button.Label.Text = fmt.Sprintf("Button %d", num)
				button.Icon.Image = icon
				button.Actions[ui.ControlTapped] = func(button *ui.Button) {
					text := fmt.Sprintf("Tapped %s", button.Label.Text)
					if num%2 == 0 {
						statusText.SetText(text)
					} else {
						// Prototype of an alert box
						statusText.SetText("Showing Alert")
						showSimpleAlert(text, "OK", func() {
							statusText.SetText("Dismissed Alert")
						})
					}
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
	cpuprofile := flag.String("cpuprofile", "", "Enable CPU Profiling to the given file")
	flag.Parse()

	if *cpuprofile != "" {
		profileOut, err := os.Create(*cpuprofile)
		if err == nil {
			defer profileOut.Close()
			err = pprof.StartCPUProfile(profileOut)
		}
		if err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
	}

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

	if err := ui.RunLoop(signalCtx, window, events); err != nil {
		log.Fatal(err)
	}
}
