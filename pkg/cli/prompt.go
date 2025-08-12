// Package cli provides reusable CLI utilities for prompts and display
package cli

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

// ConfirmAction prompts for yes/no confirmation
func ConfirmAction(label string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
		Default:   "y",
	}

	_, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrAbort {
			return false, nil
		}
		return false, fmt.Errorf("confirmation failed: %w", err)
	}

	return true, nil
}
