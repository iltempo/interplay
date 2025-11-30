package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/iltempo/interplay/commands"
	"github.com/iltempo/interplay/midi"
	"github.com/iltempo/interplay/playback"
	"github.com/iltempo/interplay/sequence"
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

	// Select MIDI port
	var portIndex int
	if len(ports) == 1 {
		// Only one port, use it automatically
		portIndex = 0
		fmt.Printf("\nUsing port %d: %s\n\n", portIndex, ports[portIndex])
	} else {
		// Multiple ports, let user choose
		fmt.Print("\n")
		rl, err := readline.New(fmt.Sprintf("Select MIDI port (0-%d): ", len(ports)-1))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating readline: %v\n", err)
			os.Exit(1)
		}
		defer rl.Close()

		input, err := rl.Readline()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}

		input = strings.TrimSpace(input)
		portIndex, err = strconv.Atoi(input)
		if err != nil || portIndex < 0 || portIndex >= len(ports) {
			fmt.Fprintf(os.Stderr, "Invalid port selection: %s\n", input)
			os.Exit(1)
		}

		fmt.Printf("Using port %d: %s\n\n", portIndex, ports[portIndex])
	}

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
	cmdHandler := commands.New(engine.GetNextPattern(), engine)

	// Read commands from stdin
	err = cmdHandler.ReadLoop(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading commands: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Goodbye!")
}
