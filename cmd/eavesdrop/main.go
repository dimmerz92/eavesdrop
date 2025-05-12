package main

import (
	"os"
	"sync"

	"github.com/spf13/pflag"
)

const (
	VERSION = "v0.1.0"
	AUTHOR  = "Andrew Weymes <andrew.weymes@sittellalab.com.au>"
)

var (
	wg = &sync.WaitGroup{}

	outFlag    = pflag.StringP("out", "o", ".", "directory output path")
	extFlag    = pflag.StringP("ext", "e", "json", "config file extension")
	helpFlag   = pflag.BoolP("help", "h", false, "prints help for a command")
	configFlag = pflag.StringP("config", "c", "eavesdrop.json", "config directory")
)

func main() {
	args := os.Args

	// run without args
	if len(args) == 1 {
		// TODO: run without any flags/args
		return
	}

	switch args[1] {
	case "init":
	// TODO: generate a config file
	case "help", "--help", "-h":
		fallthrough
	default:
		// TODO: print help
		return
	}

	wg.Wait()
}
