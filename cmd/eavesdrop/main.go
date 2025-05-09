package main

import (
	"fmt"
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
	VERSION = "v0.1.0"
	AUTHOR  = "Andrew Weymes <andrew.weymes@sittellalab.com.au>"
)

var (
	wg = &sync.WaitGroup{}

	outFlag    = pflag.StringP("out", "o", ".", "directory output path")
	extFlag    = pflag.StringP("ext", "e", "json", "config file extension")
	helpFlag   = pflag.BoolP("help", "h", false, "prints help for a command")
	configFlag = pflag.StringP("config", "c", "eavesdrop_config.json", "config directory")
)

func main() {
	args := os.Args

	// run without args
	if len(args) == 1 {
		// get config or use default
		cfg, err := config.GetConfig(*configFlag)
		if err != nil {
			cfg = config.DefaultConfig("")
			color.Yellow("warning: no config specified, using default")
		}
		// start watcher
		watcher := notify.NewWatcher(cfg)
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
			println(help)
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

var splash = fmt.Sprintf(`
███████╗ █████╗ ██╗   ██╗███████╗███████╗██████╗ ██████╗  ██████╗ ██████╗ 
██╔════╝██╔══██╗██║   ██║██╔════╝██╔════╝██╔══██╗██╔══██╗██╔═══██╗██╔══██╗
█████╗  ███████║██║   ██║█████╗  ███████╗██║  ██║██████╔╝██║   ██║██████╔╝
██╔══╝  ██╔══██║╚██╗ ██╔╝██╔══╝  ╚════██║██║  ██║██╔══██╗██║   ██║██╔═══╝ 
███████╗██║  ██║ ╚████╔╝ ███████╗███████║██████╔╝██║  ██║╚██████╔╝██║     
╚══════╝╚═╝  ╚═╝  ╚═══╝  ╚══════╝╚══════╝╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚═╝
%s
%s
Live reloading for Go apps

`, VERSION, AUTHOR)

var help = splash +
	color.YellowString("USAGE:\n") +
	color.WhiteString("\teavesdrop [COMMANDS] [OPTIONS]\n\n") +
	color.YellowString("COMMANDS:\n") +
	color.BlueString("\tinit\n") +
	color.WhiteString("\tGenerates a config file.\n\tDefaults to root directory as json if neither are specified.\n\n") +
	color.BlueString("\thelp\n") +
	color.WhiteString("\tPrints help text for eavesdrop.\n\tUse the --help or -h flags for help on commands.\n\n") +
	color.YellowString("OPTIONS:\n") +
	fmt.Sprintf("\t%s %s\n", color.MagentaString("<command>"), color.BlueString("--help, -h")) +
	color.WhiteString("\tPrints help details for the given command.\n")
