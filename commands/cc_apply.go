package commands

import (
	"fmt"
	"strconv"
)

// handleCCApply: cc-apply <cc-number>
// Applies global CC value to all steps with notes (converts transient to persistent)
func (h *Handler) handleCCApply(parts []string) error {
	if len(parts) != 2 {
		return fmt.Errorf("usage: cc-apply <cc-number>")
	}

	ccNumber, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid CC number: %s", parts[1])
	}

	// ApplyGlobalCC checks if global value is set and applies it
	if err := h.pattern.ApplyGlobalCC(ccNumber); err != nil {
		return err
	}

	// Get the value that was applied
	value, _ := h.pattern.GetGlobalCC(ccNumber)

	fmt.Printf("Applied global CC#%d (value: %d) to all steps with notes\n", ccNumber, value)
	return nil
}
