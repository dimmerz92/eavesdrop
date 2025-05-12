package notify

import "github.com/fatih/color"

// PrintError prints the given args in red.
func PrintError(s string, v ...any) {
	color.Red(s, v)
}

// PrintWarning prints the given args in yellow.
func PrintWarning(s string, v ...any) {
	color.Yellow(s, v)
}

// PrintWatching prints the given args in magenta.
func PrintWatching(s string, v ...any) {
	color.Magenta(s, v)
}

// PrintDirChange prints the given args in cyan.
func PrintDirChange(s string, v ...any) {
	color.Cyan(s, v)
}

// PrintFileChange prints the given args in green.
func PrintFileChange(s string, v ...any) {
	color.Green(s, v)
}
