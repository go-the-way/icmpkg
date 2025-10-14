package cmd

import (
	"golang.org/x/term"
	"os"
)

func getTerminalWidth() int {
	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	return width
}
