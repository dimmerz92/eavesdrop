package cli

import (
	"fmt"
	"runtime/debug"

	"github.com/fatih/color"
)

var (
	Version = func() string { v, _ := debug.ReadBuildInfo(); return v.Main.Version }()
	Author  = "Andrew Weymes <andrew.weymes@sittellalab.com.au>"
)

var Splash = fmt.Sprintf(`
███████╗ █████╗ ██╗   ██╗███████╗███████╗██████╗ ██████╗  ██████╗ ██████╗
██╔════╝██╔══██╗██║   ██║██╔════╝██╔════╝██╔══██╗██╔══██╗██╔═══██╗██╔══██╗
█████╗  ███████║██║   ██║█████╗  ███████╗██║  ██║██████╔╝██║   ██║██████╔╝
██╔══╝  ██╔══██║╚██╗ ██╔╝██╔══╝  ╚════██║██║  ██║██╔══██╗██║   ██║██╔═══╝
███████╗██║  ██║ ╚████╔╝ ███████╗███████║██████╔╝██║  ██║╚██████╔╝██║
╚══════╝╚═╝  ╚═╝  ╚═══╝  ╚══════╝╚══════╝╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚═╝
%s
%s
Live reloading for any app!

`, Version, Author)

var Help = Splash +
	fmt.Sprintf("%s eavesdrop %s %s\n\n", color.YellowString("USAGE:"), color.BlueString("[COMMAND]"), color.MagentaString("[OPTIONS]")) +
	color.YellowString("COMMANDS:\n\n") +
	fmt.Sprintf("%s %s: Generates a config file.\n", color.BlueString("init"), color.MagentaString("[options]")) +
	fmt.Sprintf("\t%s: directory to save the generated config. Defaults to [.]\n", color.MagentaString("-out")) +
	fmt.Sprintf("\t%s: the filetype to generate (json, toml, yaml). Defaults to json\n", color.MagentaString("-ext")) +
	fmt.Sprintf("\n%s: Prints the help text for eavesdrop\n\n", color.BlueString("help")) +
	fmt.Sprintf("%s: can be used without any commands\n\n", color.YellowString("OPTIONS:")) +
	fmt.Sprintf("%s: The path of the config file. Auto-detected from eavesdrop.{json,toml,yaml} in the current directory", color.MagentaString("-config"))
