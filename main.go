package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/onyx-and-iris/voicemeeter/v2"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func vmInitConnection() (*voicemeeter.Remote, *chan string, error) {
	vm, err := voicemeeter.NewRemote("banana", 20)
	if err != nil {
		return nil, nil, err
	}

	err = vm.Login()
	if err != nil {
		return nil, nil, err
	}

	events := []string{"pdirty", "mdirty"}
	vm.EventAdd(events...)
	vm_events := make(chan string)
	vm.Register(vm_events)

	return vm, &vm_events, nil
}

type PadConfig struct {
	CC      uint8    `json:"cc"`
	RGB     [3]uint8 `json:"rgb"`
	Static  bool     `json:"static" default:"false"`
	Reverse bool     `json:"reverse" default:"false"`
}

type Config struct {
	Mode string `json:"mode"`

	SourceDeviceName   string   `json:"source_device_name"`
	TargetDevicesNames []string `json:"target_devices_names"`

	PadsConfig []PadConfig `json:"pads_config"`

	VmParameterToCc map[string]uint8 `json:"vm_parameter_to_cc"`
	VmButtonToCc    map[int]uint8    `json:"vm_button_to_cc"`
}

type MidiDevice struct {
	Name    string
	InPort  drivers.In
	OutPort drivers.Out
}

func NewMidiDeviceFromName(name string, inPorts midi.InPorts, outPorts midi.OutPorts, allow_missing_ports bool) (*MidiDevice, error) {
	midiDevice := MidiDevice{
		Name: name,
	}
	for _, port := range inPorts {
		port_name := port.String()
		split := strings.Split(port_name, " ")
		port_name = strings.Join(split[:len(split)-1], " ")
		if port_name == midiDevice.Name {
			midiDevice.InPort = port
			break
		}
	}
	for _, port := range outPorts {
		port_name := port.String()
		split := strings.Split(port_name, " ")
		port_name = strings.Join(split[:len(split)-1], " ")
		if port_name == midiDevice.Name {
			midiDevice.OutPort = port
			break
		}
	}
	if !allow_missing_ports && midiDevice.InPort == nil {
		return nil, fmt.Errorf("MIDI in port %s not found", name)
	}
	if !allow_missing_ports && midiDevice.InPort == nil {
		return nil, fmt.Errorf("MIDI out port %s not found", name)
	}
	return &midiDevice, nil
}

func loadConfigFromFile(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func main() {

	config, err := loadConfigFromFile("config.json")
	if err != nil {
		log.Fatal(err)
	}

	n_pads := 8
	padsStateArray := NewPadsStateArray(n_pads, config.PadsConfig)

	ccToStateArrayIndex := make(map[uint8]int, n_pads)
	for i, padConfig := range config.PadsConfig {
		ccToStateArrayIndex[padConfig.CC] = i
	}

	drv, err := rtmididrv.New()
	if err != nil {
		log.Fatalf("Failed to initialize MIDI driver: %v", err)
	}
	defer drv.Close()

	// List available MIDI input devices
	inPorts, err := drv.Ins()
	if err != nil {
		log.Fatalf("Failed to get MIDI input ports: %v", err)
	}
	if len(inPorts) == 0 {
		log.Fatalf("No MIDI input devices found.")
	}

	// List available MIDI output devices
	outPorts, err := drv.Outs()
	if err != nil {
		log.Fatalf("Failed to get MIDI output ports: %v", err)
	}
	if len(outPorts) == 0 {
		log.Fatalf("No MIDI output devices found.")
	}

	// Print available MIDI input devices
	fmt.Println("Available MIDI input devices:")
	for i, port := range inPorts {
		fmt.Printf("%d. %s\n", i+1, port.String())
	}

	// Print available MIDI output devices
	fmt.Println("Available MIDI output devices:")
	for i, port := range outPorts {
		fmt.Printf("%d. %s\n", i+1, port.String())
	}

	sourceMidiDevice, err := NewMidiDeviceFromName(config.SourceDeviceName, inPorts, outPorts, false)
	if err != nil {
		log.Fatalf("Failed to find source MIDI device: %v", err)
	}
	targetMidiDevices := make([]*MidiDevice, len(config.TargetDevicesNames))
	for i, targetDeviceName := range config.TargetDevicesNames {
		targetMidiDevice, err := NewMidiDeviceFromName(targetDeviceName, inPorts, outPorts, false)
		if err != nil {
			log.Fatalf("Failed to find source MIDI device: %v", err)
		}
		targetMidiDevices[i] = targetMidiDevice
	}

	if err := sourceMidiDevice.InPort.Open(); err != nil {
		log.Fatalf("Failed to open MIDI input port: %v", err)
	}
	defer sourceMidiDevice.InPort.Close()

	if err := sourceMidiDevice.OutPort.Open(); err != nil {
		log.Fatalf("Failed to open MIDI input port: %v", err)
	}
	defer sourceMidiDevice.OutPort.Close()

	for _, targetMidiDevice := range targetMidiDevices {
		if err := targetMidiDevice.InPort.Open(); err != nil {
			log.Fatalf("Failed to open MIDI input port: %v", err)
		}
		defer targetMidiDevice.InPort.Close()

		if err := targetMidiDevice.OutPort.Open(); err != nil {
			log.Fatalf("Failed to open MIDI input port: %v", err)
		}
		defer targetMidiDevice.OutPort.Close()
	}

	for _, targetMidiDevice := range targetMidiDevices {
		log.Printf("Resending events from %s to %s", sourceMidiDevice.Name, targetMidiDevice.Name)
	}

	stopSend, err := midi.ListenTo(sourceMidiDevice.InPort, func(msg midi.Message, milliseconds int32) {
		if config.Mode == "switch" {
			var ch, cc, value uint8
			if msg.GetControlChange(&ch, &cc, &value) {
				stateArrayIndex, ok := ccToStateArrayIndex[cc]
				if ok && value != 0 {
					padsStateArray.Switch(stateArrayIndex)
					sourceMidiDevice.OutPort.Send(midi.SysEx(padsStateArray.ToColorChangeSysEx()))
				}
			}
		}
		for _, targetMidiDevice := range targetMidiDevices {
			// log.Printf("%s -> %s: %s", sourceMidiDevice.Name, targetMidiDevice.Name, msg.String())
			if err := targetMidiDevice.OutPort.Send(msg); err != nil {
				log.Printf("Failed to send MIDI message: %v", err)
			}
		}
	}, midi.UseSysEx())

	defer stopSend()

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	for _, targetMidiDevice := range targetMidiDevices {
		stopReturn, err := midi.ListenTo(targetMidiDevice.InPort, func(msg midi.Message, milliseconds int32) {
			// log.Printf("%s <- %s: %s", sourceMidiDevice.Name, targetMidiDevice.Name, msg.String())
			if err := sourceMidiDevice.OutPort.Send(msg); err != nil {
				log.Printf("Failed to send MIDI message: %v", err)
			}
		}, midi.UseSysEx())

		defer stopReturn()

		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			return
		}
	}

	sourceMidiDevice.OutPort.Send(midi.SysEx(padsStateArray.ToColorChangeSysEx()))

	if config.Mode == "vm" {
		vm, vm_events, err := vmInitConnection()
		log.Printf("Voicemeeter connection established")
		if err != nil {
			log.Fatal(err)
		}
		defer vm.Logout()

		go vm_events_rutine(vm, vm_events, ccToStateArrayIndex, padsStateArray, sourceMidiDevice, config)
		log.Printf("Voicemeeter events listener started")
		update_pads_from_vm("pdirty", vm, config.VmParameterToCc, config.VmButtonToCc, ccToStateArrayIndex, padsStateArray, sourceMidiDevice)
		update_pads_from_vm("mdirty", vm, config.VmParameterToCc, config.VmButtonToCc, ccToStateArrayIndex, padsStateArray, sourceMidiDevice)
	}

	select {}
}

func vm_events_rutine(vm *voicemeeter.Remote, vm_events *chan string, ccToStateArrayIndex map[uint8]int, pads_state_array *PadsStateArray, source_device *MidiDevice, config *Config) {
	events := *vm_events
	for s := range events {
		update_pads_from_vm(s, vm, config.VmParameterToCc, config.VmButtonToCc, ccToStateArrayIndex, pads_state_array, source_device)
	}
}

func update_pads_from_vm(s string, vm *voicemeeter.Remote, vm_parameter_to_cc map[string]uint8, vm_button_to_cc map[int]uint8, cc_to_state_array_index map[uint8]int, matrix *PadsStateArray, source_device *MidiDevice) {
	switch s {
	case "pdirty":
		for parameter, cc := range vm_parameter_to_cc {
			value, err := vm.GetString(parameter)
			if err != nil {
				log.Printf("Error: %s", err)
			}
			state_array_index, ok := cc_to_state_array_index[cc]
			// log.Printf("Pad %d: %v, %v", state_array_index, cc, value)
			if ok {
				matrix.GetElement(state_array_index).On = value == "1.000"
			}
		}
		source_device.OutPort.Send(midi.SysEx(matrix.ToColorChangeSysEx()))
	case "mdirty":
		for button, cc := range vm_button_to_cc {
			value := vm.Button[button].State()
			state_array_index, ok := cc_to_state_array_index[cc]
			if ok {
				matrix.GetElement(state_array_index).On = value
			}
		}
		source_device.OutPort.Send(midi.SysEx(matrix.ToColorChangeSysEx()))
	default:
	}
}
