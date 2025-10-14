package cmd

import (
	"fmt"
	"strings"
	"time"
)

func prints(ip4 string) {
	print1()
	print2(ip4)
	fmt.Println()
	print3()
	print4()
}

func print1() {
	text := "My Traceroute By Go."
	terminalWidth := getTerminalWidth()
	textWidth := len(text)
	padding := (terminalWidth - textWidth) / 2
	if padding < 0 {
		padding = 0
	}
	spaces := strings.Repeat(" ", padding)
	fmt.Printf("\033[2K\r%s%s\n", spaces, boldText(text))
}

func print2(ip4 string) {
	left := fmt.Sprintf("%s (%s) -> %s (%s)", hostname, localAddr(), target, ip4)
	right := time.Now().Format("2006-01-02T15:04:05Z0700")
	terminalWidth := getTerminalWidth()
	textWidth := len(left) + len(right)
	padding := terminalWidth - textWidth
	if padding < 0 {
		padding = 0
	}
	spaces := strings.Repeat(" ", padding)
	fmt.Printf("\033[2K\r%s%s%s\n", left, spaces, right)
}

// ***Packets*** **********Pings**********
//
// Loss%   Sent   Last   Avg   Best   Worst
// 12.2%  99999  999.0 999.0  999.0   999.0
func print3() {
	text := "Packets          Pings          "
	terminalWidth := getTerminalWidth()
	textWidth := len(text)
	padding := terminalWidth - textWidth
	if padding < 0 {
		padding = 0
	}
	spaces := strings.Repeat(" ", padding)
	fmt.Printf("\033[2K\r%s%s          %s          \n", spaces, boldText("Packets"), boldText("Pings"))
}

func print4() {
	fmt.Println(" \033[1mHost\033[0m")
}

//	Packets                Pings
//
// Loss%   Sent   Last   Avg   Best   Worst
// 12.2%  99999  999.0 999.0  999.0   999.0
func printPackets() {

}

func boldText(text string) string {
	return "\033[1m" + text + "\033[0m"
}
