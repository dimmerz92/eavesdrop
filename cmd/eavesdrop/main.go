package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/dimmerz92/eavesdrop/internal/cli"
	"github.com/dimmerz92/eavesdrop/internal/config"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if len(os.Args) == 1 {
		cli.RunEavesdrop(ctx)
		return
	}

	if os.Args[1] == "help" {
		println(cli.Help)
		return
	}

	if os.Args[1] == "init" {
		f := flag.NewFlagSet("init", flag.ContinueOnError)
		out := f.String("out", ".", "the output directory for the generated config")
		ext := f.String("ext", ".json", "the ouput format for the generated config (json, toml, yaml)")

		err := f.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}

		err = config.GenerateConfig(*out, *ext)
		if err != nil {
			panic(err)
		}
		return
	}

	if strings.HasPrefix(os.Args[1], "-") {
		cli.RunEavesdrop(ctx)
		return
	}

	panic(fmt.Sprintf("eavesdrop %s: unknown command\nRun 'eavesdrop help' for usage.", os.Args[1]))
}
