package main

import (
	"flag"
	"fmt"
	"github.com/holoplot/go-evdev"
	"hotas-to-gamepad/device"
	"hotas-to-gamepad/mapping"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"
)

var log *Logger

func main() {
	// TODO: load config from file
	deviceIds := []device.PersistentID{
		"c148e4170dd3c24665f1e342286ff51d741a0257aeeadebdc5ce3344a7c79349",
		"c08d66aba94efae239aaa86185f3fc22c84ce668ab38623af48b0f831a3312eb",
		"67ef78796bec77c6c2766f172a9d2ddab8bcdd87cef49759ed94e62bb8bb15e7",
		"4f2e1f9dd2d2b28d27363e6a748593a30c6650afdf8b5e0e7d762ebe9bb46a41",
	}
	keymap := mapping.Table{
		device.PersistentID("c148e4170dd3c24665f1e342286ff51d741a0257aeeadebdc5ce3344a7c79349"): {
			evdev.EV_ABS: {
				evdev.ABS_X: {
					Type: evdev.EV_ABS,
					Code: evdev.ABS_X,
					Mode: mapping.ModeShiftFromPositive,
				},
				evdev.ABS_Y: {
					Type: evdev.EV_ABS,
					Code: evdev.ABS_Y,
					Mode: mapping.ModeShiftFromPositive,
				},
			},
			evdev.EV_KEY: {
				evdev.BTN_TRIGGER: {
					Type: evdev.EV_ABS,
					Code: evdev.ABS_HAT2X,
					Mode: mapping.ModeExact,
				},
			},
		},
		// throttle
		device.PersistentID("c08d66aba94efae239aaa86185f3fc22c84ce668ab38623af48b0f831a3312eb"): {
			evdev.EV_ABS: {
				evdev.ABS_Y: {
					Type: evdev.EV_ABS,
					Code: evdev.ABS_HAT2X,
					Mode: mapping.ModeExact,
				},
			},
		},
		//  rudder
		//input.PersistentID("67ef78796bec77c6c2766f172a9d2ddab8bcdd87cef49759ed94e62bb8bb15e7"): {
		//	evdev.EV_ABS: {
		//		evdev.ABS_Y: {
		//			Type: evdev.EV_ABS,
		//			Code: evdev.ABS_HAT2X,
		//			Mode: mapping.ModeExact,
		//		},
		//		evdev.ABS_X: {
		//			Type: evdev.EV_ABS,
		//			Code: evdev.ABS_HAT2Y,
		//			Mode: mapping.ModeExact,
		//		},
		//	},
		//},
		// xbox controller
		device.PersistentID("4f2e1f9dd2d2b28d27363e6a748593a30c6650afdf8b5e0e7d762ebe9bb46a41"): {
			evdev.EV_ABS: {
				evdev.ABS_X: {
					Type: evdev.EV_ABS,
					Code: evdev.ABS_X,
					Mode: mapping.ModeExact,
				},
				evdev.ABS_Y: {
					Type: evdev.EV_ABS,
					Code: evdev.ABS_Y,
					Mode: mapping.ModeExact,
				},
				evdev.ABS_RX: {
					Type: evdev.EV_ABS,
					Code: evdev.ABS_RX,
					Mode: mapping.ModeExact,
				},
				evdev.ABS_RY: {
					Type: evdev.EV_ABS,
					Code: evdev.ABS_RY,
					Mode: mapping.ModeExact,
				},
				evdev.ABS_Z: {
					Type: evdev.EV_ABS,
					//Code: evdev.ABS_HAT2X,
					Code: evdev.ABS_Z,
					Mode: mapping.ModeExact,
				},
				evdev.ABS_RZ: {
					Type: evdev.EV_ABS,
					//Code: evdev.ABS_HAT2Y,
					Code: evdev.ABS_RZ,
					Mode: mapping.ModeExact,
				},
				evdev.ABS_HAT0X: {
					Type: evdev.EV_KEY,
					Code: evdev.BTN_DPAD_UP,
					Mode: mapping.ModeExact,
				},
				evdev.ABS_HAT0Y: {
					Type: evdev.EV_ABS,
					Code: evdev.ABS_HAT0Y,
					Mode: mapping.ModeExact,
				},
			},
			evdev.EV_KEY: {
				evdev.BTN_THUMBL: {
					Type: evdev.EV_KEY,
					Code: evdev.BTN_THUMBL,
					Mode: mapping.ModeExact,
				},
				evdev.BTN_THUMBR: {
					Type: evdev.EV_KEY,
					Code: evdev.BTN_THUMBR,
					Mode: mapping.ModeExact,
				},
				evdev.BTN_TL: {
					Type: evdev.EV_KEY,
					Code: evdev.BTN_TL,
					Mode: mapping.ModeExact,
				},
				evdev.BTN_TR: {
					Type: evdev.EV_KEY,
					Code: evdev.BTN_TR,
					Mode: mapping.ModeExact,
				},
				evdev.BTN_START: {
					Type: evdev.EV_KEY,
					Code: evdev.BTN_START,
					Mode: mapping.ModeExact,
				},
				evdev.BTN_SELECT: {
					Type: evdev.EV_KEY,
					Code: evdev.BTN_SELECT,
					Mode: mapping.ModeExact,
				},
				evdev.BTN_MODE: { // xbox button
					Type: evdev.EV_KEY,
					Code: evdev.BTN_MODE,
					Mode: mapping.ModeExact,
				},
				evdev.BTN_A: {
					Type: evdev.EV_KEY,
					Code: evdev.BTN_A,
					Mode: mapping.ModeExact,
				},
				evdev.BTN_B: {
					Type: evdev.EV_KEY,
					Code: evdev.BTN_B,
					Mode: mapping.ModeExact,
				},
				evdev.BTN_X: {
					Type: evdev.EV_KEY,
					Code: evdev.BTN_X,
					Mode: mapping.ModeExact,
				},
				evdev.BTN_Y: {
					Type: evdev.EV_KEY,
					Code: evdev.BTN_Y,
					Mode: mapping.ModeExact,
				},
			},
		},
	}

	// parse flags
	isDebugLogging := flag.Bool("debug", false, "enable debug logging")
	isDebugShowSyncEvents := flag.Bool("debug-show-sync-events", false, "when used with --debug, sync events will be logged")
	flag.Parse()

	// configure logging
	var logLevel LogLevel
	if *isDebugLogging {
		logLevel = LogLevelDebug
	} else {
		logLevel = LogLevelInfo
	}
	log = CreateLogger(logLevel)

	// handle os signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	// find and open configured inputs by unique id
	var inputs []*device.Device
	events := make(chan device.Event)
	errors := make(chan error)
	deviceMetas, err := device.List()
	if err != nil {
		panic(err)
	}
	for _, deviceMeta := range deviceMetas {
		fmt.Printf("Found device %s (ID %s, Path %s)\n", deviceMeta.Name, deviceMeta.PersistentID, deviceMeta.Path)
		if slices.Contains(deviceIds, deviceMeta.PersistentID) {
			device, err := deviceMeta.Open()
			if err != nil {
				panic(err)
			}
			channels := device.Listen()
			channels.Events = events
			channels.Errors = errors
			inputs = append(inputs, device)
			printDeviceInfo(device)
		}
	}

	// create virtual output
	output, err := device.Create()
	if err != nil {
		panic(err)
	}

	// throttle axis events to reduce console spam
	axisLogTimes := make(map[device.PersistentID]map[evdev.EvCode]time.Time)
	for _, input := range inputs {
		axisLogTimes[input.PersistentID()] = make(map[evdev.EvCode]time.Time)
	}

	// listen for input events
	var shouldTerminate bool
	for !shouldTerminate && len(inputs) > 0 {
		select {
		// terminate loop on exit signals
		case <-signals:
			shouldTerminate = true
			continue
		// panic on input read errors
		case err = <-errors:
			panic(err)
		// handle normal input events
		case ev := <-events:
			shouldLog :=
			// if EV_ABS, has enough time passed since the last event?
				(ev.Type != evdev.EV_ABS || time.Now().Sub(axisLogTimes[ev.Device.PersistentID()][ev.Code]).Milliseconds() >= 200) &&
					// show sync events if the flag is set
					(ev.Type != evdev.EV_SYN || *isDebugShowSyncEvents)
			typeVal := fmt.Sprintf("%d (%s)", ev.Type, ev.TypeName())
			codeVal := fmt.Sprintf("%d (%s)", ev.Code, ev.CodeName())
			if shouldLog {
				log.Debug(
					"Type: %10s\tSeq: %1d\tTime: %10d.%-6d\tCode: %-30s\tValue in: %5d",
					typeVal, ev.Sequence, ev.Time.Sec, ev.Time.Usec, codeVal, ev.Value,
				)
			}

			// remap the event and send it to the virtual output
			if ev.Type == evdev.EV_SYN {
				err := output.Send(&ev.InputEvent)
				if err != nil {
					panic(err)
				}
			} else {
				remappedEvent, ok := keymap.Remap(&ev)
				if ok {
					err := output.Send(remappedEvent)
					if err != nil {
						panic(err)
					}
					if shouldLog {
						log.Debug("\tValue out = %5d", remappedEvent.Value)
					}
				}
			}

			if shouldLog {
				log.Debug("\n")
				axisLogTimes[ev.Device.PersistentID()][ev.Code] = time.Now()
			}
		}
	}

	// close devices on loop termination
	for _, device := range inputs {
		_ = device.Close()
	}
	_ = output.Close()
}

func printDeviceInfo(device *device.Device) {
	log.Info("Input device %s (ID, %s, Path %s)\n", device.Name(), device.PersistentID(), device.Raw().Path())

	inputID, err := device.Raw().InputID()
	if err != nil {
		panic(err)
	}
	log.Info(
		"  Vendor: %d\n  Product: %d\n  Version: %d\n  Bus type: %d\n",
		inputID.Vendor, inputID.Product, inputID.Version, inputID.BusType,
	)
	absInfos, err := device.Raw().AbsInfos()
	if err != nil {
		panic(err)
	}

	// print device capabilities
	for _, evType := range device.Raw().CapableTypes() {
		log.Info("  Type %d (%s)\n", evType, evdev.TypeName(evType))
		for _, code := range device.Raw().CapableEvents(evType) {
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
}
