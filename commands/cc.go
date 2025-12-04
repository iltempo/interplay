package commands

import (
	"fmt"
	"strconv"
)

// handleCC: cc <cc-number> <value>
// Sets a global CC value that affects the entire pattern (transient, not saved)
func (h *Handler) handleCC(parts []string) error {
	if len(parts) != 3 {
		return fmt.Errorf("usage: cc <cc-number> <value>")
	}

	ccNumber, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid CC number: %s (must be 0-127)", parts[1])
	}

	value, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("invalid CC value: %s (must be 0-127)", parts[2])
	}

	// SetGlobalCC validates the CC number and value
	if err := h.pattern.SetGlobalCC(ccNumber, value); err != nil {
		return err
	}

	fmt.Printf("Set global CC#%d to %d (will take effect at next loop iteration)\n", ccNumber, value)
	return nil
}
