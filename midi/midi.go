package midi

import (
	"fmt"
	"time"

	"github.com/murdinc/go-coremidi"
)

type MidiConnection struct {
	client coremidi.Client
	input  coremidi.InputPort
	output coremidi.OutputPort
}

func New() (MidiConnection, error) {
	client, err := coremidi.NewClient("MVRD_TX7_Patcher")
	midi := MidiConnection{client: client}
	midi.initCoreMidi()
	return midi, err
}

func (midi *MidiConnection) Identity() string {

	sources, err := coremidi.AllDevices()

	if err != nil {
		fmt.Println(err)

	}

	fmt.Printf("Verbose: %v", sources)
	fmt.Printf("Client: %v", midi.client)

	//port, err := coremidi.NewOutputPort(midi.client, "MVRD_TX7_Patcher_Output")

	return ""
}

func (midi *MidiConnection) ListDestinations() {

	destinations, err := coremidi.AllDestinations()

	if err != nil {
		log("", err)

	}

	for d, destination := range destinations {
		log(fmt.Sprintf("Destination [%d] Name: %s        Manufacturer: %s", d+1, destination.Name(), destination.Manufacturer()), nil)
	}

}

func (midi *MidiConnection) ListSources() {

	sources, err := coremidi.AllSources()

	if err != nil {
		fmt.Println(err)

	}

	for s, source := range sources {
		log(fmt.Sprintf("Source [%d] Name: %s        Manufacturer: %s", s+1, source.Name(), source.Manufacturer()), nil)
	}

}

func (midi *MidiConnection) TestSend(sysex []byte) {

	destinations, _ := coremidi.AllDestinations()

	for d, destination := range destinations {
		log(fmt.Sprintf("               Destination #: %d       Name: %s        Manufacturer: %s", d+1, destination.Name(), destination.Manufacturer()), nil)

		if d == 3 {

			midi.testNotes(destination)

			sysex := coremidi.NewSysexMessage(&destination, []byte{0xF0, 0x43, 0x20, 0x09, 0xF7}, func(sysexMsg *coremidi.SysexMessage) {
				log(fmt.Sprintf("Sysex Dump Command Completed: [%X]\n", sysexMsg.Message), nil)
				return
			})
			err := sysex.Send()

			if err != nil {
				log(fmt.Sprintf("                                          failed to send sysex"), nil)
				log("", nil)
			}

		}

		//time.Sleep(100 * time.Millisecond)

	}

	time.Sleep(time.Minute)
}

func (midi *MidiConnection) testNotes(destination coremidi.Destination) {

	packet := coremidi.NewPacket([]byte{0x90, 0x3C, 100})
	_ = packet.Send(&midi.output, &destination)

	time.Sleep(100 * time.Millisecond)

	packet = coremidi.NewPacket([]byte{0x80, 0x3C, 0})
	_ = packet.Send(&midi.output, &destination)

	time.Sleep(100 * time.Millisecond)

	//

	packet = coremidi.NewPacket([]byte{0x90, 0x3E, 100})
	_ = packet.Send(&midi.output, &destination)

	time.Sleep(100 * time.Millisecond)

	packet = coremidi.NewPacket([]byte{0x80, 0x3E, 0})
	_ = packet.Send(&midi.output, &destination)

	time.Sleep(100 * time.Millisecond)

	//

	packet = coremidi.NewPacket([]byte{0x90, 0x40, 100})
	_ = packet.Send(&midi.output, &destination)

	time.Sleep(100 * time.Millisecond)

	packet = coremidi.NewPacket([]byte{0x80, 0x40, 0})
	_ = packet.Send(&midi.output, &destination)

	time.Sleep(100 * time.Millisecond)

}

func (midi *MidiConnection) Monitor() {

	sources, err := coremidi.AllSources()

	if err != nil {
		fmt.Println(err)
		return
	}

	port, err := coremidi.NewInputPort(midi.client, "MVRD_TX7_Patcher_Input", func(source coremidi.Source, value []byte) {
		log(fmt.Sprintf("source: %v manufacturer: %v value: %v valueHex: %X\n", source.Name(), source.Manufacturer(), value, value), nil)
		return
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, source := range sources {
		func(source coremidi.Source) {
			port.Connect(source)
		}(source)
	}

	ch := make(chan int)
	<-ch
}

func (midi *MidiConnection) initCoreMidi() {

	// Grab all sources
	sources, err := coremidi.AllSources()
	if err != nil {
		panic(err)
	}

	// Creates a new routine to run when something comes in on port
	input, err := coremidi.NewInputPort(midi.client, "test", func(source coremidi.Source, value []byte) {

		if value[0] == 0xF0 {
			log(fmt.Sprintf("SYSEX MESSAGE: %X	Source: %v", value, source.Name()), nil)
		} else {
			log(fmt.Sprintf("NON SYSEX MESSAGE: %X	Source: %v", value, source.Name()), nil)
		}

		return
	})
	if err != nil {
		panic(err)
	}

	// passes all sources to the port.Connect function
	for _, source := range sources {
		func(source coremidi.Source) {
			input.Connect(source)
		}(source)
	}

	//==========================

	//
	output, err := coremidi.NewOutputPort(midi.client, "test")
	if err != nil {
		panic(err)
	}

	midi.output = output
	midi.input = input
}

// Log Function
////////////////..........
func log(kind string, err error) {
	if err == nil {
		fmt.Printf("  %s\n", kind)
	} else {
		fmt.Printf("[ERROR - %s]: %s\n", kind, err)
	}
}
