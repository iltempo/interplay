package commands

import (
	"fmt"
	"strconv"
)

// handleCCClear: cc-clear <step> [cc-number]
// Clears CC automation from a step (all CC or specific CC number)
func (h *Handler) handleCCClear(parts []string) error {
	if len(parts) < 2 || len(parts) > 3 {
		return fmt.Errorf("usage: cc-clear <step> [cc-number]")
	}

	step, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid step number: %s", parts[1])
	}

	if len(parts) == 3 {
		// Clear specific CC number
		ccNumber, err := strconv.Atoi(parts[2])
		if err != nil {
			return fmt.Errorf("invalid CC number: %s", parts[2])
		}

		if err := h.pattern.ClearStepCC(step, ccNumber); err != nil {
			return err
		}

		fmt.Printf("Cleared CC#%d from step %d\n", ccNumber, step)
	} else {
		// Clear all CC automation from step
		if err := h.pattern.ClearStepCC(step, -1); err != nil {
			return err
		}

		fmt.Printf("Cleared all CC automation from step %d\n", step)
	}

	return nil
}
