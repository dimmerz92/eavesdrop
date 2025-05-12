package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/dimmerz92/eavesdrop/internal/config"
	"github.com/dimmerz92/eavesdrop/internal/notify"
	"github.com/dimmerz92/eavesdrop/internal/utils"
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
		color.Yellow(splash)

		// get the config
		cfg, err := config.GetConfig(*configFlag)
		if err != nil {
			cfg = config.DefaultConfig("")
			utils.PrintWarning("warning: no config specified, using default")
		}

		// start the notifier
		notifier := notify.NewNotifier(cfg)
		go notifier.Start()

		// run cleanup
		wg.Add(1)
		go cleanup(notifier)
		wg.Wait()
		return
	}

	switch args[1] {
	case "init":
	// TODO: generate a config file
	case "help", "--help", "-h":
		println(help)
	}
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
	color.BlueString("\tinit ") + color.MagentaString("[options]\n") +
	color.WhiteString("\tGenerates a config file.\n\n") +
	color.BlueString("\thelp\n") +
	color.WhiteString("\tPrints help text for eavesdrop.\n\tUse the --help or -h flags for help on commands.\n\n") +
	color.YellowString("OPTIONS:\n") +
	color.BlueString("\t<command> ") + color.MagentaString("--help, -h\n") +
	color.WhiteString("\tPrints help details for the given command.\n")
