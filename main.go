package main

import (
	"fmt"
	"os"

	"iltempo.de/midi-ai/commands"
	"iltempo.de/midi-ai/midi"
	"iltempo.de/midi-ai/playback"
	"iltempo.de/midi-ai/sequence"
)

func main() {
	// List available MIDI ports
	ports, err := midi.ListPorts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing MIDI ports: %v\n", err)
		os.Exit(1)
	}

	if len(ports) == 0 {
		fmt.Fprintf(os.Stderr, "No MIDI output ports found\n")
		os.Exit(1)
	}

	fmt.Println("Available MIDI ports:")
	for i, port := range ports {
		fmt.Printf("  %d: %s\n", i, port)
	}

	// For now, use the first port (later we can add selection)
	portIndex := 0
	fmt.Printf("\nUsing port %d: %s\n\n", portIndex, ports[portIndex])

	// Open MIDI output
	midiOut, err := midi.Open(portIndex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening MIDI port: %v\n", err)
		os.Exit(1)
	}
	defer midiOut.Close()

	// Create initial pattern (default: C3 on beats)
	initialPattern := sequence.New()

	// Create playback engine
	engine := playback.New(midiOut, initialPattern)

	// Start playback in background
	engine.Start()
	defer engine.Stop()

	fmt.Println("Playback started! Type 'help' for commands, 'quit' to exit.")
	fmt.Println()

	// Create command handler that modifies the "next" pattern
	cmdHandler := commands.New(engine.GetNextPattern())

	// Read commands from stdin
	err = cmdHandler.ReadLoop(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading commands: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Goodbye!")
}
