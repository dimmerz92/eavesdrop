package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/dimmerz92/eavesdrop/internal/config"
	"github.com/dimmerz92/eavesdrop/internal/notify"
	"github.com/fatih/color"
	"github.com/spf13/pflag"
)

const (
	VERSION = "0.1.0"
	AUTHOR  = "Andrew Weymes <andrew.weymes@sittellalab.com.au>"
)

var (
	wg = &sync.WaitGroup{}

	outFlag  = pflag.StringP("out", "o", ".", "directory output path")
	extFlag  = pflag.StringP("ext", "e", "json", "config file extension")
	helpFlag = pflag.BoolP("help", "h", false, "prints help for a command")
)

func main() {
	args := os.Args

	// run without args
	if len(args) == 1 {
		color.Yellow("warning: no config specified, using default")
		watcher := notify.NewWatcher(config.DefaultConfig(""))
		go watcher.Start()
		wg.Add(1)
		go cleanup(watcher)
	} else {
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

	// wait for cleanup
	wg.Wait()
}

func cleanup(w *notify.Watcher) {
	defer wg.Done()

	// listen for SIGINT & SIGTERM
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig

	color.Cyan("shutdown with signal: %s", s.String())

	// close the watcher
	w.Close()
}
