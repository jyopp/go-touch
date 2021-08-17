package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"os"
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
	display    = &ui.Display{}
	events     = &ui.EventStream{}
	background = &Background{}
	statusArea = &ui.BasicLayer{}
	statusText = &ui.TextLayer{}
)

func setStatusText(text string) {
	statusText.Text = text
	statusArea.SetNeedsDisplay()
}

func downloadBackground(button *ui.Button) {
	button.SetDisabled(true)
	button.SetNeedsDisplay()
	setStatusText("Downloading Wallpaper Image…")
	events.DisplayNeedsUpdate()

	const url = "https://news-cdn.softpedia.com/images/news2/here-are-all-iphone-and-mac-wallpapers-ever-released-by-apple-528707-3.jpg"
	if resp, err := http.Get(url); err == nil {
		defer resp.Body.Close()
		if wallpaper, err := jpeg.Decode(resp.Body); err == nil {
			imageLayer := &ui.ImageLayer{}
			imageLayer.Init(background.Bounds(), wallpaper)
			background.InsertChild(imageLayer, 0)
			setStatusText("Loaded Wallpaper")
		} else {
			setStatusText("Decode Error: " + err.Error())
		}
	} else {
		setStatusText("HTTP Error: " + err.Error())
	}
	events.DisplayNeedsUpdate()
}

func buildUI() {
	background.Init(display.Bounds(), 0xEE)

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
	for idx, rect := range buttonArea.Divide(3, 10, ui.FromTop) {
		for idx2, rect := range rect.Divide(2, 10, ui.FromLeft) {
			num := 2*idx + idx2
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

	display.AddLayer(background)
}

func main() {
	if framebuffer, err := os.OpenFile("/dev/fb1", os.O_RDWR, 0); err == nil {
		display.Init(320, 480, framebuffer)
		defer framebuffer.Close()
	} else {
		panic(err)
	}

	if eventFile, err := os.Open("/dev/input/event0"); err == nil {
		events.Init(eventFile, ui.TouchscreenCalibration{
			Left: 235, Right: 3750,
			Top: 3800, Bottom: 80,
			Weak: 180, Strong: 80,
		})
		defer eventFile.Close()
		// events.dump = true
	} else {
		panic(err)
	}

	buildUI()

	events.DispatchLoop(display)
}
