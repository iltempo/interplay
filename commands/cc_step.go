package commands

import (
	"fmt"
	"strconv"
)

// handleCCStep: cc-step <step> <cc-number> <value>
// Sets CC automation for a specific step (persistent, saved with pattern)
func (h *Handler) handleCCStep(parts []string) error {
	if len(parts) != 4 {
		return fmt.Errorf("usage: cc-step <step> <cc-number> <value>")
	}

	step, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid step number: %s", parts[1])
	}

	ccNumber, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("invalid CC number: %s (must be 0-127)", parts[2])
	}

	value, err := strconv.Atoi(parts[3])
	if err != nil {
		return fmt.Errorf("invalid CC value: %s (must be 0-127)", parts[3])
	}

	// SetStepCC validates step, CC number, and value
	if err := h.pattern.SetStepCC(step, ccNumber, value); err != nil {
		return err
	}

	fmt.Printf("Set step %d CC#%d to %d\n", step, ccNumber, value)
	return nil
}
