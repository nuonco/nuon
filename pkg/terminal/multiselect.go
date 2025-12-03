package terminal

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/term"
)

const (
	keyUp        = 65
	keyDown      = 66
	keySpace     = 32
	keyEnter     = 13
	keyCtrlC     = 3
	keyEscape    = 27
	keyCtrlD     = 4
	cursorUp     = "\033[A"
	cursorDown   = "\033[B"
	clearLine    = "\033[2K"
	cursorStart  = "\r"
	hideCursor   = "\033[?25l"
	showCursor   = "\033[?25h"
	clearScreen  = "\033[2J"
	moveCursorTo = "\033[H"
)

// MultiSelect provides an interactive checkbox-style multi-select interface
// Users can navigate with arrow keys, toggle selections with space, and confirm with enter
func MultiSelect(prompt string, options []string, currentSelections []string) ([]string, error) {
	// Validate input
	if len(options) == 0 {
		return []string{}, nil
	}

	// Initialize selection state
	selected := make(map[string]bool)
	for _, sel := range currentSelections {
		selected[sel] = true
	}

	cursor := 0
	totalOptions := len(options)

	// Save terminal state
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer func() {
		fmt.Print(showCursor) // Always show cursor when done
		term.Restore(int(os.Stdin.Fd()), oldState)
	}()

	// Setup signal handler to restore terminal on interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Print(showCursor)
		term.Restore(int(os.Stdin.Fd()), oldState)
		os.Exit(1)
	}()
	defer signal.Stop(sigCh)

	// Hide cursor and print initial UI
	fmt.Print(hideCursor)
	fmt.Printf("\r\n%s:\r\n", prompt)
	fmt.Print("Use ↑/↓ to navigate, SPACE to toggle, ENTER to confirm\r\n")
	printOptions(options, selected, cursor)

	// Read key presses
	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}

		// Handle different key combinations
		if n == 1 {
			switch buf[0] {
			case keySpace:
				// Toggle selection
				option := options[cursor]
				selected[option] = !selected[option]
				redrawOptions(options, selected, cursor, totalOptions)

			case keyEnter:
				// Confirm and exit
				clearDisplay(totalOptions)
				fmt.Print("\r\n") // Move to next line after clearing
				return getSelectedOptions(options, selected), nil

			case keyCtrlC, keyCtrlD:
				// Cancel
				clearDisplay(totalOptions)
				fmt.Print("Cancelled\r\n")
				return currentSelections, nil
			}
		} else if n == 3 && buf[0] == keyEscape && buf[1] == '[' {
			// Arrow key
			switch buf[2] {
			case keyUp:
				if cursor > 0 {
					cursor--
					redrawOptions(options, selected, cursor, totalOptions)
				}
			case keyDown:
				if cursor < totalOptions-1 {
					cursor++
					redrawOptions(options, selected, cursor, totalOptions)
				}
			}
		}
	}
}

// printOptions prints all options with checkboxes
func printOptions(options []string, selected map[string]bool, cursor int) {
	for i, opt := range options {
		checkbox := "[ ]"
		if selected[opt] {
			checkbox = "[✓]"
		}

		marker := "  "
		if i == cursor {
			marker = "> "
		}

		fmt.Printf("%s%s %s\r\n", marker, checkbox, opt)
	}
}

// redrawOptions clears and redraws the options list
func redrawOptions(options []string, selected map[string]bool, cursor int, totalOptions int) {
	// Move cursor up to the start of the list
	for i := 0; i < totalOptions; i++ {
		fmt.Print(cursorUp)
	}

	// Clear each line and redraw
	for i, opt := range options {
		fmt.Print(cursorStart + clearLine)

		checkbox := "[ ]"
		if selected[opt] {
			checkbox = "[✓]"
		}

		marker := "  "
		if i == cursor {
			marker = "> "
		}

		fmt.Printf("%s%s %s\r\n", marker, checkbox, opt)
	}
}

// clearDisplay clears the display area
func clearDisplay(lines int) {
	for i := 0; i < lines; i++ {
		fmt.Print(cursorUp)
	}
	for i := 0; i < lines; i++ {
		fmt.Print(cursorStart + clearLine + "\r\n")
	}
	// Move back up
	for i := 0; i < lines; i++ {
		fmt.Print(cursorUp)
	}
	fmt.Print(cursorStart)
}

// getSelectedOptions returns the list of selected options
func getSelectedOptions(options []string, selected map[string]bool) []string {
	result := []string{}
	for _, opt := range options {
		if selected[opt] {
			result = append(result, opt)
		}
	}
	return result
}

// RenderSelections shows the final selections in a readable format
func RenderSelections(selections []string) string {
	if len(selections) == 0 {
		return "none"
	}
	return strings.Join(selections, ", ")
}
