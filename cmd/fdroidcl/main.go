// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"mvdan.cc/fdroidcl/basedir"
)

const cmdName = "fdroidcl"

var version = "v0.3.1"

func errExit(format string, a ...interface{}) {
	fmt.Fprintf(stderr, format, a...)
	os.Exit(1)
}

func subdir(dir, name string) string {
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(p, 0755); err != nil {
		errExit("Could not create dir '%s': %v\n", p, err)
	}
	return p
}

var (
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr

	testBasedir = ""
)

func mustCache() string {
	if testBasedir != "" {
		return subdir(testBasedir, "cache")
	}
	dir := basedir.Cache()
	if dir == "" {
		errExit("Could not determine cache dir\n")
	}
	return subdir(dir, cmdName)
}

func mustData() string {
	if testBasedir != "" {
		return subdir(testBasedir, "data")
	}
	dir := basedir.Data()
	if dir == "" {
		errExit("Could not determine data dir\n")
	}
	return subdir(dir, cmdName)
}

func configPath() string {
	return filepath.Join(mustData(), "config.json")
}

type repo struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

type userConfig struct {
	Repos []repo `json:"repos"`
}

var config = userConfig{
	Repos: []repo{
		{
			ID:      "f-droid",
			URL:     "https://f-droid.org/repo",
			Enabled: true,
		},
		{
			ID:      "f-droid-archive",
			URL:     "https://f-droid.org/archive",
			Enabled: false,
		},
	},
}

func readConfig() {
	f, err := os.Open(configPath())
	if err != nil {
		return
	}
	defer f.Close()
	fileConfig := userConfig{}
	if err := json.NewDecoder(f).Decode(&fileConfig); err == nil {
		config = fileConfig
	}
}

// A Command is an implementation of a go command
// like go build or go fix.
type Command struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(args []string) error

	// UsageLine is the one-line usage message.
	// The first word in the line is taken to be the command name.
	UsageLine string

	// Short is the short description.
	Short string
}

// Name returns the command's name: the first word in the usage line.
func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (c *Command) usage(flagSet *flag.FlagSet) {
	fmt.Fprintf(stderr, "Usage: %s %s [-h]\n", cmdName, c.UsageLine)
	anyFlags := false
	flagSet.VisitAll(func(f *flag.Flag) { anyFlags = true })
	if anyFlags {
		fmt.Fprintf(stderr, "\nAvailable options:\n")
		flagSet.PrintDefaults()
	}
	os.Exit(2)
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(stderr, "Usage: %s [-h] <command> [<args>]\n\n", cmdName)
		fmt.Fprintf(stderr, "Available commands:\n")
		maxUsageLen := 0
		for _, c := range commands {
			if len(c.UsageLine) > maxUsageLen {
				maxUsageLen = len(c.UsageLine)
			}
		}
		for _, c := range commands {
			fmt.Fprintf(stderr, "   %s%s  %s\n", c.UsageLine,
				strings.Repeat(" ", maxUsageLen-len(c.UsageLine)), c.Short)
		}
		fmt.Fprintf(stderr, "\nA specific version of an app can be selected by following the appid with an colon (:) and the version code of the app to select.\n")
		fmt.Fprintf(stderr, "\nUse %s <command> -h for more info\n", cmdName)
	}
}

// Commands lists the available commands.
var commands = []*Command{
	cmdUpdate,
	cmdSearch,
	cmdShow,
	cmdList,
	cmdDevices,
	cmdDownload,
	cmdInstall,
	cmdUpgrade,
	cmdUninstall,
	cmdDefaults,
	cmdVersion,
}

var cmdVersion = &Command{
	UsageLine: "version",
	Short:     "Print version information",
	Run: func(args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("no arguments allowed")
		}
		fmt.Fprintln(stdout, version)
		return nil
	},
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		flag.Usage()
		os.Exit(2)
	}

	cmdName := args[0]
	for _, cmd := range commands {
		if cmd.Name() != cmdName {
			continue
		}
		readConfig()
		if err := cmd.Run(args[1:]); err != nil {
			errExit("%s: %v\n", cmdName, err)
		}
		return
	}

	switch cmdName {
	default:
		fmt.Fprintf(stderr, "Unrecognised command '%s'\n\n", cmdName)
		flag.Usage()
		os.Exit(2)
	}
}
