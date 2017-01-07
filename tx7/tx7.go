package tx7

import (
	"errors"
	"fmt"
	"time"

	"github.com/murdinc/portmidi"
	"github.com/murdinc/terminal"
)

var (
	tx7 *TX7
)

// TX7 represents a device with an input and output MIDI stream.
type TX7 struct {
	inputDevice  portmidi.DeviceId
	outputDevice portmidi.DeviceId
	inputStream  *portmidi.Stream
	outputStream *portmidi.Stream
}

func New(input portmidi.DeviceId, output portmidi.DeviceId) (*TX7, error) {
	var err error
	var inStream, outStream *portmidi.Stream
	if inStream, err = portmidi.NewInputStream(input, 1024); err != nil {
		return nil, err
	}
	if outStream, err = portmidi.NewOutputStream(output, 1024, 0); err != nil {
		return nil, err
	}
	return &TX7{inputDevice: input, outputDevice: output, inputStream: inStream, outputStream: outStream}, nil
}

func (t *TX7) Open() error {
	var err error
	var inStream, outStream *portmidi.Stream
	if inStream, err = portmidi.NewInputStream(t.inputDevice, 1024); err != nil {
		return err
	}
	if outStream, err = portmidi.NewOutputStream(t.outputDevice, 1024, 0); err != nil {
		return err
	}

	t.inputStream = inStream
	t.outputStream = outStream

	return nil

}

func (t *TX7) Close() error {
	return portmidi.Terminate()
}

// Read reads messages from the input stream. It returns max 64 messages for each read.
func (t *TX7) Read() (events []portmidi.Event, err error) {
	var evts []portmidi.Event

	t.Open()

	if evts, err = t.inputStream.Read(1024); err != nil {
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

func (t *TX7) Upload(sysex []byte) {

	err := portmidi.Initialize()
	if err != nil {
		return
	}

	fmt.Sprintf("%d %d", t.inputDevice, t.outputDevice)

	t.Open()

	err = t.outputStream.WriteSysExBytes(portmidi.Time(), sysex)
	if err != nil {
		log("WriteSysEx", err)
	}

	t.TestNotes()

}

func (t *TX7) DownloadVoice(callback func(data []byte)) {
	var sysexMessage []byte

	t.Open()

	// Set up Listener
	ch := t.Listen()

	sysexRecieving := false
	sysexRequest := []byte{0xF0, 0x43, 0x20, 0x00, 0x00, 0xF7} // 1 voice

	t.outputStream.WriteSysExBytes(portmidi.Time(), sysexRequest)

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

}

func (t *TX7) DownloadBank(callback func(data []byte)) {

	var sysexMessage []byte

	t.Open()

	// Set up Listener
	ch := t.Listen()

	sysexRecieving := false
	sysexRequest := []byte{0xF0, 0x43, 0x20, 0x09, 0x00, 0xF7} // 32 voices

	t.outputStream.WriteSysEx(portmidi.Time(), string(sysexRequest))

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
}

func (t *TX7) TestNotes() {

	// note on events to play C# minor chord
	err := t.outputStream.WriteShort(0x90, 60, 100)
	if err != nil {
		log("testNotes", err)
	}
	t.outputStream.WriteShort(0x90, 64, 100)
	t.outputStream.WriteShort(0x90, 67, 100)

	time.Sleep(time.Second / 4)

	// note off events
	t.outputStream.WriteShort(0x80, 60, 100)
	t.outputStream.WriteShort(0x80, 64, 100)
	t.outputStream.WriteShort(0x80, 67, 100)

}

func Discover() (input portmidi.DeviceId, output portmidi.DeviceId, err error) {

	err = portmidi.Initialize()
	if err != nil {
		return
	}

	if portmidi.CountDevices() < 1 {
		err = errors.New("No MIDI Device Detected!")
		return
	}

	for i := 0; i < portmidi.CountDevices(); i++ {
		info := portmidi.GetDeviceInfo(portmidi.DeviceId(i))

		if info.IsInputAvailable {
			terminal.Response(fmt.Sprintf("[ %d ]		Input		%s", i+1, info.Name))
		}
		if info.IsOutputAvailable {
			terminal.Prompt(fmt.Sprintf("[ %d ]		Output		%s", i+1, info.Name))
		}
	}

	inputDevice := terminal.PromptInt("Please select the MIDI INPUT device:", portmidi.CountDevices()+1) - 1
	outputDevice := terminal.PromptInt("Please select the MIDI OUTPUT device:", portmidi.CountDevices()+1) - 1

	input = portmidi.DeviceId(inputDevice)
	output = portmidi.DeviceId(outputDevice)

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
