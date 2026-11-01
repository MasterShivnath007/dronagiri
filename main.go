// Command dronagiri puts every command-line tool on your Mac in one place.
// what it is, how big it is, when it arrived, and what depends on it.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MasterShivnath007/dronagiri/internal/model"
	"github.com/MasterShivnath007/dronagiri/internal/runner"
	"github.com/MasterShivnath007/dronagiri/internal/source"
	"github.com/MasterShivnath007/dronagiri/internal/ui"
)

const version = "1.0.0"

func main() {
	args := os.Args[1:]

	if len(args) > 0 {
		switch args[0] {
		case "list", "ls", "installed":
			runList(args[1:])
			return
		case "logo", "emblem":
			fmt.Println(ui.Emblem(120))
			return
		case "version", "--version", "-v":
			fmt.Printf("dronagiri %s\n", version)
			return
		case "help", "--help", "-h":
			printHelp()
			return
		case "--json":
			runList(args)
			return
		}
	}

	runTUI()
}

func runTUI() {
	program := tea.NewProgram(ui.New(runner.New()), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "dronagiri:", err)
		os.Exit(1)
	}
}

func runList(args []string) {
	flags := flag.NewFlagSet("list", flag.ExitOnError)
	asJSON := flags.Bool("json", false, "print machine-readable JSON")
	showAll := flags.Bool("all", false, "include dependencies")
	onlyDeps := flags.Bool("deps", false, "show only dependencies")
	onlyBroken := flags.Bool("broken", false, "show only packages whose files are missing")
	noColor := flags.Bool("no-color", false, "disable colour")
	_ = flags.Parse(args)

	r := runner.New()
	inv := source.LoadAll(r, nil)

	if len(inv.Errors) > 0 && !*asJSON {
		for _, e := range inv.Errors {
			fmt.Fprintln(os.Stderr, "warning:", e.Error())
		}
	}

	var tools []*model.Tool
	for _, tool := range inv.Tools {
		var keep bool
		switch {
		case *onlyBroken:
			keep = tool.Broken
		case *onlyDeps:
			keep = !tool.UserInstalled
		case *showAll:
			keep = true
		default:
			keep = tool.UserInstalled
		}
		if keep {
			tools = append(tools, tool)
		}
	}

	if *asJSON {
		type row struct {
			*model.Tool
			RequiredBy []string `json:"required_by"`
		}
		rows := make([]row, 0, len(tools))
		for _, tool := range tools {
			rows = append(rows, row{Tool: tool, RequiredBy: inv.RequiredBy[tool.ID]})
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(rows); err != nil {
			fmt.Fprintln(os.Stderr, "dronagiri:", err)
			os.Exit(1)
		}
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PACKAGE\tVERSION\tSIZE\tINSTALLED\tSOURCE\tNOTES")
	var total int64
	for _, tool := range tools {
		total += tool.SizeBytes
		notes := ""
		switch {
		case tool.Broken:
			notes = "files missing"
		case tool.Protected:
			notes = "protected"
		case tool.IsDependency():
			if names := inv.RequiredBy[tool.ID]; len(names) > 0 {
				notes = "needed by " + strings.Join(names, ", ")
			} else {
				notes = "unused dependency"
			}
		case tool.Outdated():
			notes = "update: " + tool.Latest
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			tool.Name,
			tool.Version,
			model.HumanBytes(tool.SizeBytes),
			model.HumanDate(tool.InstalledAt),
			tool.Source.Label(),
			notes,
		)
	}
	_ = w.Flush()

	if !*noColor {
		fmt.Fprintf(os.Stdout, "\n%d packages · %s\n", len(tools), model.HumanBytes(total))
	}
}

func printHelp() {
	fmt.Print(`Dronagiri — every tool on your Mac, lifted into one place

Usage:
  dronagiri                 interactive browser (recommended)
  dronagiri list            list packages you installed yourself
  dronagiri list --all      include dependencies
  dronagiri list --deps     only dependencies, showing what needs them
  dronagiri list --broken   manager entries whose files are gone
  dronagiri list --json     machine-readable output
  dronagiri logo            print the emblem on its own
  dronagiri version         print the version

Inside the interactive browser:
  ↑ ↓ / j k   move            space   select
  d           remove          u       update
  /           search          tab     switch view
  a           select all      r       reload
  ?           help            q       quit
`)
}
