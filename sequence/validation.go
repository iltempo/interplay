package sequence

import "fmt"

// ValidateCC checks if a CC number and value are within valid MIDI range (0-127)
func ValidateCC(ccNumber, value int) error {
	if ccNumber < 0 || ccNumber > 127 {
		return fmt.Errorf("CC number must be 0-127, got %d", ccNumber)
	}
	if value < 0 || value > 127 {
		return fmt.Errorf("CC value must be 0-127, got %d", value)
	}
	return nil
}
