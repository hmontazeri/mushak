package utils

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

// ConfirmDanger prompts the user to confirm a dangerous action
func ConfirmDanger(message, confirmText string) (bool, error) {
	prompt := promptui.Prompt{
		Label: fmt.Sprintf("%s\nType '%s' to confirm", message, confirmText),
	}

	result, err := prompt.Run()
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(result) == confirmText, nil
}

// Confirm prompts the user for yes/no confirmation
func Confirm(message string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     message,
		IsConfirm: true,
	}

	result, err := prompt.Run()
	if err != nil {
		// promptui returns an error for "no" responses
		return false, nil
	}

	return strings.ToLower(result) == "y" || strings.ToLower(result) == "yes", nil
}

// PromptString prompts the user for a string input
func PromptString(label, defaultValue string) (string, error) {
	prompt := promptui.Prompt{
		Label:   label,
		Default: defaultValue,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(result), nil
}
