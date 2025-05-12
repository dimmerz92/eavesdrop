package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/dimmerz92/eavesdrop/internal/notify"
	"github.com/fatih/color"
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

func cleanup(n *notify.Notifier) {
	defer wg.Done()

	// listen for SIGINT and SIGTERM
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// block and wait for a signal
	s := <-sig

	color.Cyan("shutdown with signal: %s", s.String())
	n.Stop()
}
