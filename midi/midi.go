package midi

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/murdinc/portmidi"
)

var (
	tx7 *TX7
)

// TX7 represents a device with an input and output MIDI stream.
type TX7 struct {
	inputStream  *portmidi.Stream
	outputStream *portmidi.Stream
}

func Open() (*TX7, error) {
	input, output, err := discover()
	if err != nil {
		return nil, err
	}

	var inStream, outStream *portmidi.Stream
	if inStream, err = portmidi.NewInputStream(input, 5000); err != nil {
		return nil, err
	}
	if outStream, err = portmidi.NewOutputStream(output, 5000, 0); err != nil {
		return nil, err
	}
	return &TX7{inputStream: inStream, outputStream: outStream}, nil
}

// Read reads messages from the input stream. It returns max 64 messages for each read.
func (t *TX7) Read() (events []portmidi.Event, err error) {
	var evts []portmidi.Event
	if evts, err = t.inputStream.Read(5000); err != nil {
		return
	}

	events = append(events, evts...)
	return
}

// Listen listens the input stream for messages.
func (t *TX7) Listen() <-chan portmidi.Event {
	ch := make(chan portmidi.Event)
	go func(tx7 *TX7, ch chan portmidi.Event) {
		for {
			// sleep for a while before the new polling tick,
			// otherwise operation is too intensive and blocking
			time.Sleep(10 * time.Millisecond)
			events, err := tx7.Read()
			if err != nil {
				continue
			}
			for i := range events {
				ch <- events[i]
			}
		}
	}(t, ch)
	return ch
}

func Upload(sysex []byte) {

	fmt.Println("TestWriteSysex")

	var err error

	err = portmidi.Initialize()
	if err != nil {
		log("Initialization", err)
	}

	if tx7, err = Open(); err != nil {
		log("error while initializing connection to tx7", err)
	}

	// listen button toggles
	ch := tx7.Listen()
	go func() {

		for {
			event := <-ch

			log(fmt.Sprintf("IN: Message: [%X] Status: [%X]", event.Message, event.Message.Status()), nil)

		}
	}()

	//sysexMessage := []byte{0xF0, 0x43, 0x20, 0x09, 0xF7}

	err = tx7.outputStream.WriteSysEx(portmidi.Time(), string(sysex))
	if err != nil {
		log("WriteSysEx", err)
	}

	time.Sleep(time.Minute)

	portmidi.Terminate()

}

func DownloadVoice(callback func(data []byte)) {
	var err error

	err = portmidi.Initialize()
	if err != nil {
		log("Initialization", err)
	}

	if tx7, err = Open(); err != nil {
		log("error while initializing connection to tx7", err)
	}

	var sysexMessage []byte

	// Set up Listener
	ch := tx7.Listen()

	sysexRecieving := false
	sysexRequest := []byte{0xF0, 0x43, 0x20, 0x00, 0x00, 0xF7} // 1 voice

	tx7.outputStream.WriteSysEx(portmidi.Time(), string(sysexRequest))

Loop:
	for {
		event := <-ch

		// Start or continue recieving a sysex message
		if sysexRecieving == true || event.Message[0] == 0xF0 {
			sysexRecieving = true

			for i := 0; i < len(event.Message); i++ {

				sysexMessage = append(sysexMessage, event.Message[i])

				if event.Message[i] == 0xF7 {
					sysexRecieving = false

					callback(sysexMessage)

					break Loop

				}
			}

		}

	}

	portmidi.Terminate()

}

func DownloadBank(callback func(data []byte)) {
	var err error

	err = portmidi.Initialize()
	if err != nil {
		log("Initialization", err)
	}

	if tx7, err = Open(); err != nil {
		log("error while initializing connection to tx7", err)
	}

	var sysexMessage []byte

	// Set up Listener
	ch := tx7.Listen()

	sysexRecieving := false
	sysexRequest := []byte{0xF0, 0x43, 0x20, 0x09, 0x00, 0xF7} // 32 voices

	tx7.outputStream.WriteSysEx(portmidi.Time(), string(sysexRequest))

Loop:
	for {
		event := <-ch

		// Start or continue recieving a sysex message
		if sysexRecieving == true || event.Message[0] == 0xF0 {
			sysexRecieving = true

			for i := 0; i < len(event.Message); i++ {

				sysexMessage = append(sysexMessage, event.Message[i])

				if event.Message[i] == 0xF7 {
					sysexRecieving = false

					callback(sysexMessage)

					break Loop

				}
			}

		}

	}

	portmidi.Terminate()
}

func (t *TX7) TestNotes() {

	// note on events to play C# minor chord
	err := t.outputStream.WriteShort(0x90, 60, 100)
	if err != nil {
		log("testNotes", err)
	}
	t.outputStream.WriteShort(0x90, 64, 100)
	t.outputStream.WriteShort(0x90, 67, 100)

	// notes will be sustained for 2 seconds
	time.Sleep(2 * time.Second)

	// note off events
	t.outputStream.WriteShort(0x80, 60, 100)
	t.outputStream.WriteShort(0x80, 64, 100)
	t.outputStream.WriteShort(0x80, 67, 100)

}

func discover() (input portmidi.DeviceId, output portmidi.DeviceId, err error) {
	in := -1
	out := -1
	for i := 0; i < portmidi.CountDevices(); i++ {
		info := portmidi.GetDeviceInfo(portmidi.DeviceId(i))
		if strings.Contains(info.Name, " 1") {
			if info.IsInputAvailable {
				in = i
				log(fmt.Sprintf("Input: %s", info.Name), nil)
			}
			if info.IsOutputAvailable {
				out = i
				log(fmt.Sprintf("Output: %s", info.Name), nil)
			}
		}
	}
	if in == -1 || out == -1 {
		err = errors.New("No Device Connected!")
	} else {
		input = portmidi.DeviceId(in)
		output = portmidi.DeviceId(out)
	}
	return
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

/*
func New() (MidiConnection, error) {
	fmt.Println("New")
	client, err := coremidi.NewClient("MVRD_TX7_Patcher")
	fmt.Println("New2")
	midi := MidiConnection{client: client}
	midi.initCoreMidi()
	return midi, err
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

	port, err := coremidi.NewInputPort(midi.client, "MVRD_TX7_Patcher_Input", func() {
		//log(fmt.Sprintf("source: %v manufacturer: %v value: %v valueHex: %X\n", source.Name(), source.Manufacturer(), value, value), nil)
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
	fmt.Println("initCoreMidi1")
	sources, err := coremidi.AllSources()
	fmt.Println("initCoreMidi2")
	if err != nil {
		panic(err)
	}

	// Creates a new routine to run when something comes in on port
	input, err := coremidi.NewInputPort(midi.client, "test", func() {

		//fmt.Printf(" MAIN FUNC: %X", message)

		//if value[0] == 0xF0 {
		//log(fmt.Sprintf("SYSEX MESSAGE: %X	Source: %v", value, source.Name()), nil)
		//} else {
		//log(fmt.Sprintf("NON SYSEX MESSAGE: %X	Source: %v", value, source.Name()), nil)
		//}

		return
	})
	if err != nil {
		panic(err)
	}

	// passes all sources to the port.Connect function
	for _, source := range sources {
		func(source coremidi.Source) {
			fmt.Println("Connect")
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


*/
