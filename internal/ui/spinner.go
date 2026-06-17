package ui

import (
	"time"

	"github.com/pterm/pterm"
)

// Global instance to manage the current active spinner
var activeSpinner *pterm.SpinnerPrinter

// StartSpinner shows an animated loader with a custom message
func StartSpinner(message string) {
	// If a spinner is already running, stop it first
	if activeSpinner != nil {
		activeSpinner.Stop()
	}

	// Setup the spinner using PTerm's professional theme
	spinner, _ := pterm.DefaultSpinner.
		WithSequence("|", "/", "-", "\\").
		WithDelay(time.Millisecond * 100).
		Start(message)

	activeSpinner = spinner
}

// UpdateSpinner changes the message of the currently running loader
func UpdateSpinner(message string) {
	if activeSpinner != nil {
		activeSpinner.UpdateText(message)
	}
}

// StopSpinnerSuccess stops the loader and marks it with a green checkmark
func StopSpinnerSuccess(message string) {
	if activeSpinner != nil {
		activeSpinner.Success(message)
		activeSpinner = nil
	}
}

// StopSpinnerFail stops the loader and marks it with a red cross
func StopSpinnerFail(message string) {
	if activeSpinner != nil {
		activeSpinner.Fail(message)
		activeSpinner = nil
	}
}

// ActionMessage simply prints a beautifully formatted step message without a loader
func ActionMessage(action string, target string) {
	pterm.Info.Printf("%s: %s\n", pterm.LightCyan(action), pterm.White(target))
}

// PrintTable displays server lists or status in a clean, aligned table format
func PrintTable(headers []string, data [][]string) {
	tableData := [][]string{headers}
	tableData = append(tableData, data...)

	pterm.DefaultTable.
		WithHasHeader().
		WithBoxed().
		WithData(tableData).
		Render()
}
