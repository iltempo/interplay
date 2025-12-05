package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/chzyer/readline"
	"github.com/iltempo/interplay/commands"
	"github.com/iltempo/interplay/midi"
	"github.com/iltempo/interplay/playback"
	"github.com/iltempo/interplay/sequence"
	"github.com/mattn/go-isatty"
)

// isTerminal returns true if stdin is a terminal (TTY)
func isTerminal() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

// processBatchInput reads and executes commands from reader
// Returns (success, shouldExit) where success indicates no errors occurred
// and shouldExit indicates if an explicit exit command was found
func processBatchInput(reader io.Reader, handler *commands.Handler) (bool, bool) {
	scanner := bufio.NewScanner(reader)
	hadErrors := false
	shouldExit := false

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Print comments (for user visibility)
		if strings.HasPrefix(line, "#") {
			fmt.Println(line)
			continue
		}

		// Check for explicit exit command
		if strings.ToLower(line) == "exit" || strings.ToLower(line) == "quit" {
			shouldExit = true
			continue
		}

		// Echo command for progress feedback
		fmt.Println(">", line)

		// Process command
		if err := handler.ProcessCommand(line); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			hadErrors = true
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		return false, shouldExit
	}

	return !hadErrors, shouldExit
}

func main() {
	// Parse command-line flags
	scriptFile := flag.String("script", "", "execute commands from file")
	flag.Parse()
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
	// Auto-select port 0 in batch mode (script file or piped input)
	inBatchMode := *scriptFile != "" || !isTerminal()

	if len(ports) == 1 || inBatchMode {
		// Only one port, or batch mode - use port 0 automatically
		portIndex = 0
		fmt.Printf("\nUsing port %d: %s\n\n", portIndex, ports[portIndex])
	} else {
		// Multiple ports in interactive mode, let user choose
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
	initialPattern := sequence.New(sequence.DefaultPatternLength)

	// Create playback engine
	engine := playback.New(midiOut, initialPattern)

	// Start playback in background
	engine.Start()
	defer engine.Stop()

	// Setup cleanup function for graceful shutdown
	cleanup := func() {
		engine.Stop()
		midiOut.Close()
	}

	// Setup signal handler for Ctrl+C to ensure clean shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down gracefully...")
		cleanup()
		os.Exit(0)
	}()

	fmt.Println("Playback started! Type 'help' for commands, 'quit' to exit.")
	fmt.Println()

	// Create command handler that modifies the "next" pattern
	cmdHandler := commands.New(engine.GetNextPattern(), engine)

	// Handle script file mode
	if *scriptFile != "" {
		// Open script file
		f, err := os.Open(*scriptFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening script file: %v\n", err)
			os.Exit(2)
		}
		defer f.Close()

		// Process script file
		success, shouldExit := processBatchInput(f, cmdHandler)

		// Exit with appropriate code if exit command present or on error
		if shouldExit {
			cleanup()
			if success {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		}
		// Otherwise continue running with playback
		fmt.Println("\nScript completed. Playback continues. Press Ctrl+C to exit.")
		select {} // Block forever, playback goroutine keeps running
	}

	// Determine input mode based on stdin
	if isTerminal() {
		// Interactive mode (existing behavior)
		err = cmdHandler.ReadLoop(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading commands: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Batch mode (piped input)
		success, shouldExit := processBatchInput(os.Stdin, cmdHandler)

		// Exit with appropriate code if exit command present
		if shouldExit {
			cleanup()
			if success {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		}

		// Continue running with playback loop active (performance tool paradigm)
		// User can stop with Ctrl+C
		fmt.Println("\nBatch commands completed. Playback continues. Press Ctrl+C to exit.")
		select {} // Block forever, playback goroutine keeps running
	}

	fmt.Println("Goodbye!")
}
