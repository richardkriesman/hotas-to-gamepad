package main

import (
	_ "embed"
	"flag"
	"fmt"
	"github.com/holoplot/go-evdev"
	"github.com/richardkriesman/hotas-to-gamepad/config"
	"github.com/richardkriesman/hotas-to-gamepad/device"
	"os"
	"os/signal"
	"syscall"
	"text/template"
	"time"
)

var log *Logger

//go:embed templates/help.tmpl
var usageTemplateText string

type Flags struct {
	Debug               bool
	DebugShowSyncEvents bool
}

func main() {
	// parse flags
	debug := flag.Bool("debug", false, "enable debug logging")
	debugShowSyncEvents := flag.Bool("debug-show-sync-events", false, "when used with --debug, sync events will be logged")
	flag.Usage = func() {
		if err := template.Must(template.New("usage").Parse(usageTemplateText)).Execute(os.Stdout, struct {
			ImageName string
		}{
			ImageName: os.Args[0],
		}); err != nil {
			panic(err)
		}
		flag.PrintDefaults()
	}
	flag.Parse()
	command := flag.Arg(0)
	if command == "" {
		flag.Usage()
		os.Exit(1)
	}
	flags := Flags{
		Debug:               *debug,
		DebugShowSyncEvents: *debugShowSyncEvents,
	}

	// configure logging
	var logLevel LogLevel
	if flags.Debug {
		logLevel = LogLevelDebug
	} else {
		logLevel = LogLevelInfo
	}
	log = CreateLogger(logLevel)

	// route subcommand
	switch command {
	case "list":
		commandList()
	case "remap":
		commandRemap(flags)
	default:
		log.Error("Unknown command: %s", command)
		os.Exit(1)
	}
}

func commandList() {
	deviceMetas, err := device.List()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Available devices:")
	for _, deviceMeta := range deviceMetas {
		fmt.Printf("  %s (ID %s, Path %s)\n", deviceMeta.Name, deviceMeta.PersistentID, deviceMeta.Path)
		dev, err := deviceMeta.Open()
		if err != nil {
			fmt.Printf("    ERROR: Failed to open: %s\n", err)
			continue
		}
		printDeviceInfo(dev)
		_ = dev.Close()
	}
}

func commandRemap(flags Flags) {
	events := make(chan device.Event)
	errors := make(chan error)

	// handle os signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	// load config
	configPath := flag.Arg(1)
	if configPath == "" {
		configPath = "config.yaml"
	}
	inputConfig, err := config.Load(configPath)
	if err != nil {
		panic(err)
	}
	keymap := inputConfig.ToMappingTable()

	// list devices and select those that match the config
	var inputs []*device.InputDevice
	deviceMetas, err := device.List()
	if err != nil {
		panic(err)
	}
	for _, deviceMeta := range deviceMetas {
		_, ok := inputConfig.Inputs[deviceMeta.PersistentID]
		if ok {
			dev, err := deviceMeta.Open()
			if err != nil {
				panic(err)
			}
			channels := dev.Listen()
			channels.Events = events
			channels.Errors = errors
			inputs = append(inputs, dev)
			printDeviceInfo(dev)
		}
	}

	// create virtual output
	outputConfig := device.Gamepad
	output, err := device.Create(outputConfig)
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
		case event := <-events:
			shouldLog :=
				// if EV_ABS, has enough time passed since the last event?
				(event.Type != evdev.EV_ABS || time.Now().Sub(axisLogTimes[event.Device.PersistentID()][event.Code]).Milliseconds() >= 200) &&
					// show sync events if the flag is set
					(event.Type != evdev.EV_SYN || flags.DebugShowSyncEvents)
			typeVal := fmt.Sprintf("%d (%s)", event.Type, event.TypeName())
			codeVal := fmt.Sprintf("[%s (%d)]", event.CodeName(), event.Code)
			if shouldLog {
				log.Debug(
					"Type: %10s\tSeq: %1d\tTime: %10d.%-6d\t %30s %6d ===> ",
					typeVal, event.Sequence, event.Time.Sec, event.Time.Usec, codeVal, event.Value,
				)
			}

			// remap the event and send it to the virtual output
			if event.Type == evdev.EV_SYN {
				// forward sync without modification
				if err := output.Send(&event.InputEvent); err != nil {
					panic(err)
				}
			} else {
				// remap all other events
				remappedEvent, ok := keymap.Remap(&event, outputConfig)
				if ok {
					if err := output.Send(remappedEvent); err != nil {
						panic(err)
					}
					if shouldLog {
						codeVal = fmt.Sprintf("[%s (%d)]", remappedEvent.CodeName(), remappedEvent.Code)
						log.Debug("%-6d %-30s", remappedEvent.Value, codeVal)
					}
				} else if shouldLog {
					log.Debug("!")
				}
			}

			if shouldLog {
				log.Debug("\n")
				axisLogTimes[event.Device.PersistentID()][event.Code] = time.Now()
			}
		}
	}

	// close devices on loop termination
	for _, dev := range inputs {
		_ = dev.Close()
	}
	_ = output.Close()
}

func printDeviceInfo(device *device.InputDevice) {
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
