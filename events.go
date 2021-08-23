package fbui

import (
	"encoding/binary"
	"fmt"
	"image"
	"os"
	"syscall"
)

const (
	BTN_TOUCH = 0x14a

	ABS_X        = 0x00
	ABS_Y        = 0x01
	ABS_Z        = 0x02
	ABS_PRESSURE = 0x18
)

type EventStream struct {
	DeviceFile *os.File
	Events     chan TouchEvent
	dump       bool
}

type inputEvent struct {
	Time  syscall.Timeval
	Type  uint16
	Code  uint16
	Value int32
}

type TouchEvent struct {
	image.Point
	Pressed  bool
	Pressure int
}

func (es *EventStream) Init(deviceFile *os.File) {
	es.DeviceFile = deviceFile
	es.Events = make(chan TouchEvent, 100)
}

func (es *EventStream) inputReadLoop() {
	var currentEvent TouchEvent
	var e inputEvent

	for {
		if err := binary.Read(es.DeviceFile, binary.LittleEndian, &e); err != nil {
			return
		}

		if e.Time.Sec == 0 {
			continue
		}

		if es.dump {
			fmt.Printf("Event %+v\n", e)
		}

		switch e.Type {
		case 0:
			es.Events <- currentEvent
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
