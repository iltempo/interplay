package commands

import (
	"fmt"
	"sort"
)

// ccEntry represents a single CC automation entry for display
type ccEntry struct {
	step     int
	ccNumber int
	value    int
}

// handleCCShow: cc-show
// Displays all active CC automation across all steps in a table format
func (h *Handler) handleCCShow(parts []string) error {
	if len(parts) != 1 {
		return fmt.Errorf("usage: cc-show")
	}

	// Collect all CC automation data
	var entries []ccEntry

	patternLen := h.pattern.Length()
	for step := 1; step <= patternLen; step++ {
		// Get step data and iterate directly over its CC map
		// This is O(n) instead of O(n × 128) where n = number of actual CC automations
		stepData, err := h.pattern.GetStep(step)
		if err != nil {
			continue // Skip invalid steps
		}

		// Iterate only over CC values that are actually set
		for ccNum, value := range stepData.CCValues {
			entries = append(entries, ccEntry{
				step:     step,
				ccNumber: ccNum,
				value:    value,
			})
		}
	}

	// Check if there's any CC automation
	if len(entries) == 0 {
		fmt.Println("No CC automation configured")
		return nil
	}

	// Sort by step, then by CC number
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].step != entries[j].step {
			return entries[i].step < entries[j].step
		}
		return entries[i].ccNumber < entries[j].ccNumber
	})

	// Display table header
	fmt.Println("CC Automation:")
	fmt.Println("  Step  CC#  Value")
	fmt.Println("  ----  ---  -----")

	// Display entries
	for _, entry := range entries {
		fmt.Printf("  %4d  %3d  %5d\n", entry.step, entry.ccNumber, entry.value)
	}

	fmt.Printf("\nTotal: %d CC automation(s) across %d step(s)\n", len(entries), countUniqueSteps(entries))
	return nil
}

// countUniqueSteps counts how many unique steps have CC automation
func countUniqueSteps(entries []ccEntry) int {
	steps := make(map[int]bool)
	for _, entry := range entries {
		steps[entry.step] = true
	}
	return len(steps)
}
