package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
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

	// TODO: run without args
	if len(args) == 1 {
		runEavesdrop()
		return
	}

	switch args[1] {
	case "help":
		println(help)
		return

	case "init":
		outputDir := flag.String("out", ".", "the desired output directory")
		ext := flag.String("ext", ".json", "the desired format ('.json', '.toml', '.yaml')")
		flag.Parse()

		err := eavesdrop.GenerateConfig(*outputDir, *ext)
		if err != nil {
			color.Red("error: %v", err)
		}
		return

	default:
		fmt.Printf("eavesdrop %s: unknown command\nRun 'eavesdrop help' for usage.", args[1])
		os.Exit(1)
	}
}

func runEavesdrop() {
	println(splash)

	path := flag.String("config", ".", "the path to the config file")

	if *path == "." {
		*path = ".eavesdrop.json"
	}

	config, err := eavesdrop.GetConfig(*path)
	if err != nil {
		color.Red("error: no config found: %v", err)
	}

	manager, err := eavesdrop.NewEventManager(config)
	if err != nil {
		color.Red("error: %v", err)
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

	// block and wait for a signal
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
