package main

import (
	"image"
	"image/color"

	"github.com/jyopp/go-touch"
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
	AlertBoxConfig.Border = color.RGBA{A: 0x30}
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
	touch.BasicLayer
	Title, Message touch.TextLayer
	Buttons        []*touch.Button
	Border         color.Color
}

func (alert *AlertBox) Init() {
	alert.BasicLayer.Radius = 8
	alert.Background = AlertBoxConfig.Background
	alert.Border = AlertBoxConfig.Border

	alert.Title.Init(image.Rectangle{}, AlertBoxConfig.TitleFont.Name, AlertBoxConfig.TitleFont.Size)
	alert.Title.Gravity = touch.GravityCenter

	alert.Message.Init(image.Rectangle{}, AlertBoxConfig.MessageFont.Name, AlertBoxConfig.MessageFont.Size)
	alert.Message.Gravity = touch.GravityCenter

	alert.AddChild(&alert.Title, &alert.Message)
	// Don't forget this, else method overrides will fail
	alert.Self = alert
}

func (alert *AlertBox) AddButton(label string, action func()) *touch.Button {
	button := &touch.Button{}

	button.Init(image.Rectangle{}, AlertBoxConfig.ButtonFont.Name, AlertBoxConfig.ButtonFont.Size)
	button.Label.SetText(label)
	button.Label.Gravity = touch.GravityCenter
	button.Colors.Normal.Text = AlertBoxConfig.ButtonFont.Color
	button.Colors.Normal.Background = AlertBoxConfig.ButtonBackground
	button.Actions[touch.ControlTapped] = func(button *touch.Button) {
		alert.RemoveFromParent()
		action()
	}

	alert.Buttons = append(alert.Buttons, button)
	alert.AddChild(button)
	alert.LayoutInParent()
	return button
}

func (alert *AlertBox) NaturalSize() image.Point {
	const padding = 10
	padCount := 4 // 2 for top/bottom, and double spacing around message

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
	var buttonHeight = 0
	if buttonCount := len(alert.Buttons); buttonCount > 2 {
		// Vertical stack
		buttonHeight = 48 * buttonCount
		padCount += buttonCount - 1
	} else if buttonCount > 0 {
		// Buttons will be laid out left/right
		buttonHeight = 60
	}

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
func (alert *AlertBox) SetParent(parent touch.Layer) {
	alert.BasicLayer.SetParent(parent)
	alert.LayoutInParent()
}

func (alert *AlertBox) LayoutInParent() {
	if alert.Parent() == nil {
		return
	}
	parentArea := touch.Layout(alert.Parent().Frame().Inset(10))
	layout := touch.Layout(parentArea.Aligned(alert.NaturalSize(), touch.GravityCenter))
	alert.SetFrame(layout.Rectangle)
	layout = layout.InsetBy(10, 10)

	// If the title is
	if titleSize := alert.Title.NaturalSize(); titleSize.X > 1 {
		alert.Title.SetFrame(layout.Slice(titleSize.Y, 10, touch.FromTop).Rectangle)
	} else {
		alert.Title.SetFrame(image.Rectangle{})
	}
	if buttonCount := len(alert.Buttons); buttonCount > 2 {
		// Place the buttons in reverse order, slicing bottom-up
		for idx := buttonCount; idx > 0; idx-- {
			alert.Buttons[idx-1].SetFrame(layout.Slice(48, 10, touch.FromBottom).Rectangle)
		}
	} else if buttonCount > 0 {
		for idx, rect := range layout.Slice(60, 10, touch.FromBottom).Divide(buttonCount, 10, touch.FromRight) {
			alert.Buttons[idx].SetFrame(rect.Rectangle)
		}
	}
	// All remaining space is used for the message.
	alert.Message.SetFrame(layout.Rectangle)
}

func (alert *AlertBox) OpaqueRect() image.Rectangle {
	return alert.BasicLayer.OpaqueRect().Inset(1)
}

func (alert *AlertBox) DrawIn(ctx touch.DrawingContext) {
	// This could be more efficient, but it's flexible and the fill operations should be fast
	const borderWidth = 2
	ctx.Fill(alert.Rectangle, alert.Border, alert.Radius)
	ctx.Fill(alert.Rectangle.Inset(borderWidth), alert.Background, alert.Radius-borderWidth)
}
