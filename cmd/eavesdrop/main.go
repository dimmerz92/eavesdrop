package main

import (
	"os"

	"github.com/dimmerz92/eavesdrop/internal/config"
	"github.com/fatih/color"
	"github.com/spf13/pflag"
)

const (
	VERSION = "0.1.0"
	AUTHOR  = "Andrew Weymes <andrew.weymes@sittellalab.com.au>"
)

var (
	outFlag  = pflag.StringP("out", "o", ".", "directory output path")
	extFlag  = pflag.StringP("ext", "e", "json", "config file extension")
	helpFlag = pflag.BoolP("help", "h", false, "prints help for a command")
)

func main() {
	args := os.Args

	// TODO: Run with no args
	if len(args) == 1 {
		return
	}

	switch args[1] {
	case "init": // generate a config file
		pflag.Parse()
		if *helpFlag {
			// print init command help
		} else if err := config.GenerateConfig(*outFlag, *extFlag); err != nil {
			color.Red("error generating config file: %v", err)
		} else {
			color.Green("config file generated")
		}

	case "help", "--help", "-h":
		fallthrough

	default:
		// TODO: print help text
		return
	}
}
