package main

import (
	"image"
	"image/color"

	ui "github.com/jyopp/fbui"
)

var AlertBoxConfig struct {
	Background, Border, ButtonBackground color.Color

	TitleFont, MessageFont, ButtonFont struct {
		Name  string
		Size  float64
		Color color.Color
	}
}

func init() {
	AlertBoxConfig.Background = color.White
	// Dim shadow (R/G/B are zero, which is black)
	AlertBoxConfig.Border = color.RGBA{A: 0x22}
	// Light, Cool Gray buttons on white background by default
	AlertBoxConfig.ButtonBackground = color.RGBA{R: 0xE0, G: 0xE0, B: 0xF0, A: 0xFF}

	AlertBoxConfig.ButtonFont.Size = 15.0
	AlertBoxConfig.ButtonFont.Color = color.Gray{0x33}

	AlertBoxConfig.TitleFont.Size = 11.0
	AlertBoxConfig.TitleFont.Color = color.Black

	AlertBoxConfig.MessageFont.Size = 15.0
	AlertBoxConfig.MessageFont.Color = color.Black
}

type AlertBox struct {
	ui.BasicLayer
	Title, Message ui.TextLayer
	Buttons        []*ui.Button
	Border         color.Color
}

func (alert *AlertBox) Init() {
	alert.BasicLayer.Radius = 8
	alert.Background = AlertBoxConfig.Background
	alert.Border = AlertBoxConfig.Border

	alert.Title.Init(image.Rectangle{}, AlertBoxConfig.TitleFont.Name, AlertBoxConfig.TitleFont.Size)
	alert.Title.Gravity = ui.GravityCenter

	alert.Message.Init(image.Rectangle{}, AlertBoxConfig.MessageFont.Name, AlertBoxConfig.MessageFont.Size)
	alert.Message.Gravity = ui.GravityCenter

	alert.AddChild(&alert.Title, &alert.Message)
	// Don't forget this, else method overrides will fail
	alert.Self = alert
}

func (alert *AlertBox) AddButton(label string, action func()) *ui.Button {
	button := &ui.Button{}

	button.Init(image.Rectangle{}, AlertBoxConfig.ButtonFont.Name, AlertBoxConfig.ButtonFont.Size)
	button.Label.SetText(label)
	button.Label.Gravity = ui.GravityCenter
	button.Colors.Normal.Text = AlertBoxConfig.ButtonFont.Color
	button.Colors.Normal.Background = AlertBoxConfig.ButtonBackground
	button.Actions[ui.ControlTapped] = func(button *ui.Button) {
		alert.RemoveFromParent()
		action()
	}

	alert.Buttons = append(alert.Buttons, button)
	alert.AddChild(button)
	alert.LayoutInParent()
	return button
}

func (alert *AlertBox) NaturalSize() image.Point {
	// len(buttons)-1 + 2 (top+bottom) + 2 (extra space for message)
	padCount := len(alert.Buttons) + 3
	titleSize := alert.Title.NaturalSize()
	if titleSize.X > 1 {
		padCount++
	} else {
		titleSize.Y = 0
	}
	messageSize := alert.Message.NaturalSize()
	if messageSize.X > 1 {
		padCount++
	} else {
		messageSize.Y = 0
	}
	buttonHeight := 48 * len(alert.Buttons)

	size := image.Point{
		X: 280,
		Y: titleSize.Y + messageSize.Y + buttonHeight + 10*padCount,
	}
	if size.Y < 180 {
		size.Y = 180
	}
	return size
}

// Alert automatically sizes and centers itself when added to a parent view.
func (alert *AlertBox) SetParent(parent ui.Layer) {
	alert.BasicLayer.SetParent(parent)
	alert.LayoutInParent()
}

func (alert *AlertBox) LayoutInParent() {
	if alert.Parent() == nil {
		return
	}
	parentArea := ui.Layout(alert.Parent().Frame().Inset(10))
	layout := ui.Layout(parentArea.Aligned(alert.NaturalSize(), ui.GravityCenter))
	alert.SetFrame(layout.Rectangle)
	layout = layout.InsetBy(10, 10)

	// If the title is
	if titleSize := alert.Title.NaturalSize(); titleSize.X > 1 {
		alert.Title.SetFrame(layout.Slice(titleSize.Y, 10, ui.FromTop).Rectangle)
	} else {
		alert.Title.SetFrame(image.Rectangle{})
	}
	// Place the buttons in reverse order, slicing bottom-up
	for idx := len(alert.Buttons); idx > 0; idx-- {
		alert.Buttons[idx-1].SetFrame(layout.Slice(48, 10, ui.FromBottom).Rectangle)
	}
	// All remaining space is used for the message.
	alert.Message.SetFrame(layout.Rectangle)
}

func (alert *AlertBox) OpaqueRect() image.Rectangle {
	return alert.BasicLayer.OpaqueRect().Inset(1)
}

func (alert *AlertBox) DrawIn(ctx ui.DrawingContext) {
	// This could be more efficient, but it's flexible and the fill operations should be fast
	const borderWidth = 1
	ctx.Fill(alert.Rectangle, alert.Border, alert.Radius)
	ctx.Fill(alert.Rectangle.Inset(borderWidth), alert.Background, alert.Radius-borderWidth)
}
