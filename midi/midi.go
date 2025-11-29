package midi

import (
	"fmt"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // auto-register RtMIDI driver
)

// Output represents a MIDI output connection
type Output struct {
	port drivers.Out
	send func(msg midi.Message) error
}

// ListPorts returns a list of available MIDI output port names
func ListPorts() ([]string, error) {
	ports := midi.GetOutPorts()
	names := make([]string, len(ports))
	for i, port := range ports {
		names[i] = port.String()
	}
	return names, nil
}

// Open opens a MIDI output port by index
func Open(portIndex int) (*Output, error) {
	port, err := midi.OutPort(portIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get MIDI port %d: %w", portIndex, err)
	}

	send, err := midi.SendTo(port)
	if err != nil {
		return nil, fmt.Errorf("failed to create sender: %w", err)
	}

	return &Output{
		port: port,
		send: send,
	}, nil
}

// Close closes the MIDI output port
func (o *Output) Close() error {
	return o.port.Close()
}

// NoteOn sends a MIDI Note On message
// note: MIDI note number (0-127, where C4=60)
// velocity: note velocity (0-127)
// channel: MIDI channel (0-15, where 0 = channel 1)
func (o *Output) NoteOn(channel, note, velocity uint8) error {
	return o.send(midi.NoteOn(channel, note, velocity))
}

// NoteOff sends a MIDI Note Off message
func (o *Output) NoteOff(channel, note uint8) error {
	return o.send(midi.NoteOff(channel, note))
}
