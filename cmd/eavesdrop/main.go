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

var (
	VERSION string
	AUTHOR  = "Andrew Weymes <andrew.weymes@sittellalab.com.au>"
)

var (
	wg = &sync.WaitGroup{}

	outF    = pflag.StringP("out", "o", ".", "directory output path")
	extF    = pflag.StringP("ext", "e", "json", "config file extension")
	helpF   = pflag.BoolP("help", "h", false, "prints help for a command")
	configF = pflag.StringP("config", "c", ".eavesdrop.json", "config directory")
)

func main() {
	args := os.Args

	// run without args
	if len(args) == 1 {
		runEavesdrop()
		return
	}

	pflag.Parse()

	switch args[1] {
	// generate a config file
	case "init":
		if *helpF {
			println(initHelp)
			return
		}

		err := config.GenerateConfig(*outF, *extF)
		if err != nil {
			utils.PrintError("error generating config: %v", err)
		} else {
			color.Green("config file generated to: %s", *outF)
		}

	// print base help
	case "help", "--help", "-h":
		println(help)

	// run eavesdrop with flags
	default:
		runEavesdrop()
	}
}

func runEavesdrop() {
	println(splash)

	// get the config
	cfg, err := config.GetConfig(*configF)
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
	color.WhiteString("\teavesdrop ") +
	color.BlueString("[COMMANDS] ") + color.MagentaString("[OPTIONS]\n\n") +
	color.YellowString("COMMANDS:\n") +
	color.BlueString("\tinit ") + color.MagentaString("[options]\n") +
	color.WhiteString("\tGenerates a config file.\n\n") +
	color.BlueString("\thelp\n") +
	color.WhiteString("\tPrints help text for eavesdrop.\n") +
	color.WhiteString("\tUse --help or -h flags for help on commands.\n\n") +
	color.YellowString("OPTIONS:\n") +
	color.MagentaString("\t--config, -c\n") +
	color.WhiteString("\tThe directory containing the config file.\n") +
	color.WhiteString("\tDefaults to project root if not supplied.\n\n") +
	color.BlueString("\t<command> ") + color.MagentaString("--help, -h\n") +
	color.WhiteString("\tPrints help details for the given command.\n")

var initHelp = splash +
	color.YellowString("USAGE:\n") +
	color.WhiteString("\teavescrop init [OPTIONS]\n\n") +
	color.YellowString("OPTIONS:\n") +
	color.MagentaString("\t--out, -o\n") +
	color.WhiteString("\tThe output directory for the generated config.\n\n") +
	color.MagentaString("\t--ext, -e\n") +
	color.WhiteString("\tThe extension (filetype) for the config.\n") +
	color.WhiteString("\tjson, toml, and yaml are supported.\n\n")
