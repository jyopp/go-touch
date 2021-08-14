package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"os"
)

const (
	BTN_TOUCH = 0x14a

	ABS_X        = 0x00
	ABS_Y        = 0x01
	ABS_Z        = 0x02
	ABS_PRESSURE = 0x18
)

type EventStream struct {
	// Digitzer values for screen corners, and for weak / strong press
	DeviceFile  *os.File
	TouchEvents chan TouchEvent
	dump        bool
}

type TouchEvent struct {
	image.Point
	Pressed  bool
	Pressure int
}

func NewEventStream(deviceFile *os.File) *EventStream {
	return &EventStream{
		DeviceFile:  deviceFile,
		TouchEvents: make(chan TouchEvent, 100),
	}
}

func (es *EventStream) EventLoop() {
	var currentEvent TouchEvent
	var e InputEvent

	for {
		if err := binary.Read(es.DeviceFile, binary.LittleEndian, &e); err != nil {
			return
		}
		// println(len(rawEvents), "events")
		if es.dump {
			fmt.Printf("Event %+v\n", e)
		}

		if e.Time.Sec == 0 {
			continue
		}
		switch e.Type {
		case 0:
			es.TouchEvents <- currentEvent
		case 1:
			// Button event
			if e.Code == BTN_TOUCH {
				currentEvent.Pressed = e.Value > 0
			}
		case 3:
			// State event
			switch e.Code {
			case ABS_X:
				currentEvent.X = int(e.Value)
			case ABS_Y:
				currentEvent.Y = int(e.Value)
			case ABS_PRESSURE:
				currentEvent.Pressure = int(e.Value)
			}
		}
	}
}
