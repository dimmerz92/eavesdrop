package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/dimmerz92/eavesdrop"
	"github.com/fatih/color"
)

var (
	VERSION string
	AUTHOR  = "Andrew Weymes <andrew.weymes@sittellalab.com.au>"

	wg = sync.WaitGroup{}
)

func main() {
	args := os.Args

	if len(args) == 1 {
		runEavesdrop()
		return
	}

	switch {
	case args[1] == "help":
		println(help)

	case args[1] == "init":
		f := flag.NewFlagSet("init", flag.ContinueOnError)
		outputDir := f.String("out", ".", "the desired output directory")
		ext := f.String("ext", ".json", "the desired format ('.json', '.toml', '.yaml')")
		err := f.Parse(args[2:])
		if err != nil {
			color.Red("error: %v", err)
			os.Exit(1)
		}

		err = eavesdrop.GenerateConfig(*outputDir, *ext)
		if err != nil {
			color.Red("error: %v", err)
		}

	case strings.HasPrefix(args[1], "-"):
		runEavesdrop()

	default:
		fmt.Printf("eavesdrop %s: unknown command\nRun 'eavesdrop help' for usage.", args[1])
		os.Exit(1)
	}
}

func runEavesdrop() {
	println(splash)

	path := flag.String("config", ".eavesdrop.json", "the path to the config file")
	flag.Parse()

	config, err := eavesdrop.GetConfig(*path)
	if err != nil {
		color.Red("error: %v", err)
		os.Exit(1)
	}

	manager, err := eavesdrop.NewEventManager(config)
	if err != nil {
		color.Red("error: %v", err)
		os.Exit(1)
	}

	go manager.Start()

	wg.Add(1)
	go cleanup(manager)

	wg.Wait()
}

func cleanup(manager *eavesdrop.EventManager) {
	defer wg.Done()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig

	color.Cyan("shutdown with signal: %s", s.String())
	manager.Stop()
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
Live reloading for any app!

`, VERSION, AUTHOR)

var help = splash +
	color.YellowString("USAGE:\n") +
	color.WhiteString("\teavesdrop ") +
	color.BlueString("[COMMAND] ") + color.MagentaString("[OPTIONS]\n\n") +
	color.YellowString("COMMANDS:\n") +
	color.BlueString("\tinit ") + color.MagentaString("[options]\n") +
	color.WhiteString("\tGenerates a config file.\n") +
	color.MagentaString("\t-out") + color.WhiteString(" directory to save the generated config. Defaults to .\n") +
	color.MagentaString("\t-ext") + color.WhiteString(" the filetype to generate (.json, .toml, .yaml). Defaults to .json\n\n") +
	color.BlueString("\thelp\n") +
	color.WhiteString("\tPrints help text for eavesdrop.\n\n") +
	color.YellowString("OPTIONS:\n") +
	color.WhiteString("The following options can be used when running without a command:\n") +
	color.MagentaString("\t-config\n") +
	color.WhiteString("\tThe path of the config file. Defaults to ./.eavesdrop.json\n")
