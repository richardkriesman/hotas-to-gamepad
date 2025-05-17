package main

import (
	"flag"
	"fmt"
	"github.com/holoplot/go-evdev"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	keymap := RemapTable{
		evdev.EV_ABS: {
			evdev.ABS_X: {
				Type: evdev.EV_ABS,
				Code: evdev.ABS_X,
				Mode: RemapShiftFromPositiveOnly,
			},
			evdev.ABS_Y: {
				Type: evdev.EV_ABS,
				Code: evdev.ABS_Y,
				Mode: RemapShiftFromPositiveOnly,
			},
		},
		evdev.EV_KEY: {
			evdev.BTN_TRIGGER: {
				Type: evdev.EV_ABS,
				Code: evdev.ABS_HAT2X,
				Mode: RemapExact,
			},
		},
	}

	// parse flags
	isDebugLogging := flag.Bool("debug", false, "enable debug logging")
	flag.Parse()
	path := flag.Arg(0)

	// configure logging
	var logLevel LogLevel
	if *isDebugLogging {
		logLevel = LogLevelDebug
	} else {
		logLevel = LogLevelInfo
	}
	log := CreateLogger(logLevel)

	// handle os signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	// open joystick device
	joystick, err := OpenInputDevice(path)
	if err != nil {
		panic(err)
	}
	log.Info("Input device %s (%s)\n", joystick.Name(), path)

	// print device info
	id, err := joystick.RawDevice().InputID()
	if err != nil {
		panic(err)
	}
	log.Info(
		"  Vendor: %d\n  Product: %d\n  Version: %d\n  Bus type: %d\n",
		id.Vendor, id.Product, id.Version, id.BusType,
	)
	absInfos, err := joystick.RawDevice().AbsInfos()
	if err != nil {
		panic(err)
	}

	// print device capabilities
	for _, evType := range joystick.RawDevice().CapableTypes() {
		log.Info("  Type %d (%s)\n", evType, evdev.TypeName(evType))
		for _, code := range joystick.RawDevice().CapableEvents(evType) {
			// print code name and number
			log.Info("    Code %d (%s)\n", code, evdev.CodeName(evType, code))

			// print axis params if this is an axis
			if evType == evdev.EV_ABS {
				info := absInfos[code]
				log.Info(
					"      Axis Value = %d\tFlat = %d\tMinimum = %d\tMaximum = %d\tFuzz = %d\tResolution = %d\n",
					info.Value, info.Flat, info.Minimum, info.Maximum, info.Fuzz, info.Resolution,
				)
			}
		}
	}

	// create virtual gamepad
	gamepad, err := CreateGamepad()
	if err != nil {
		panic(err)
	}

	// listen for input events
	var shouldTerminate bool
	axisLogTimes := make(map[evdev.EvCode]time.Time)
	for !shouldTerminate {
		select {
		// terminate loop on exit signals
		case <-signals:
			shouldTerminate = true
			continue
		// panic on joystick read errors
		case err = <-joystick.Errors:
			panic(err)
		// handle normal joystick events
		case ev := <-joystick.Events:
			shouldLog := ev.Type != evdev.EV_ABS || time.Now().Sub(axisLogTimes[ev.Code]).Milliseconds() >= 200
			typeVal := fmt.Sprintf("%d (%s)", ev.Type, ev.TypeName())
			codeVal := fmt.Sprintf("%d (%s)", ev.Code, ev.CodeName())
			if shouldLog {
				log.Debug(
					"Type: %10s\tSeq: %1d\tTime: %10d.%-6d\tCode: %-30s\tValue in: %5d",
					typeVal, ev.Sequence, ev.Time.Sec, ev.Time.Usec, codeVal, ev.Value,
				)
			}

			// remap the event and send it to the virtual controller
			remappedEvent, ok := keymap.Remap(&ev.InputEvent, absInfos)
			if ok {
				err := gamepad.Send(remappedEvent)
				if err != nil {
					panic(err)
				}
				if shouldLog {
					log.Debug("\tValue out = %5d", remappedEvent.Value)
				}
			}

			if shouldLog {
				log.Debug("\n")
				axisLogTimes[ev.Code] = time.Now()
			}
		}
	}

	_ = joystick.Close()
	_ = gamepad.Close()
}
