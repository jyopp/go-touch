package main

import (
	"context"
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"

	ui "github.com/jyopp/fbui"
)

const (
	DefaultFont       = "Raleway-Regular.ttf"
	DefaultButtonFont = "Raleway-Medium.ttf"
	DefaultBoldFont   = "Raleway-SemiBold.ttf"
)

func init() {
	Resources.RegisterFont(DefaultFont)
	Resources.RegisterFont(DefaultButtonFont)
	Resources.RegisterFont(DefaultBoldFont)

	AlertBoxConfig.TitleFont.Name = DefaultBoldFont
	AlertBoxConfig.MessageFont.Name = DefaultFont
	AlertBoxConfig.ButtonFont.Name = DefaultButtonFont
}

var (
	window     = &ui.Window{}
	events     = &ui.EventStream{}
	background = &Background{}
	statusArea = &ui.BasicLayer{}
	statusText = &ui.TextLayer{}
)

func styleDefaultAlertButton(button *ui.Button) {
	button.Label.SetFont(DefaultBoldFont, AlertBoxConfig.ButtonFont.Size)
	button.Label.Color = color.Black
	button.Colors.Normal.Background = color.RGBA{R: 0xC0, G: 0xD0, B: 0xFF, A: 0xFF}
	button.StateDidChange()
}

func showSimpleAlert(message, buttontext string, done func()) {
	alert := &AlertBox{}
	alert.Init()

	alert.Title.Text = ""
	alert.Message.Text = message
	// Test out a 'default' button color
	styleDefaultAlertButton(alert.AddButton(buttontext, done))
	alert.AddButton("Cancel", done)

	// Alert will size itself and lay out when added to parent
	window.AddChild(alert)
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
		rowStart := 3 * idx
		for idx, rect := range rect.Divide(3, 10, ui.FromLeft) {
			num := rowStart + idx

			button := &ui.Button{}
			button.Init(rect.Rectangle, DefaultButtonFont, 15.0)
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
