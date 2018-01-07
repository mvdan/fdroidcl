// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mvdan.cc/fdroidcl/basedir"
)

const cmdName = "fdroidcl"

var version = "v0.3.1"

func errExit(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}

func subdir(dir, name string) string {
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(p, 0755); err != nil {
		errExit("Could not create dir '%s': %v\n", p, err)
	}
	return p
}

func mustCache() string {
	dir := basedir.Cache()
	if dir == "" {
		errExit("Could not determine cache dir\n")
	}
	return subdir(dir, cmdName)
}

func mustData() string {
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

	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet
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

func (c *Command) Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s %s [-h]\n", cmdName, c.UsageLine)
	anyFlags := false
	c.Flag.VisitAll(func(f *flag.Flag) { anyFlags = true })
	if anyFlags {
		fmt.Fprintf(os.Stderr, "\nAvailable options:\n")
		c.Flag.PrintDefaults()
	}
	os.Exit(2)
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-h] <command> [<args>]\n\n", cmdName)
		fmt.Fprintf(os.Stderr, "Available commands:\n")
		maxUsageLen := 0
		for _, c := range commands {
			if len(c.UsageLine) > maxUsageLen {
				maxUsageLen = len(c.UsageLine)
			}
		}
		for _, c := range commands {
			fmt.Fprintf(os.Stderr, "   %s%s  %s\n", c.UsageLine,
				strings.Repeat(" ", maxUsageLen-len(c.UsageLine)), c.Short)
		}
		fmt.Fprintf(os.Stderr, "\nA specific version of an app can be selected by following the appid with an colon (:) and the version code of the app to select.\n")
		fmt.Fprintf(os.Stderr, "\nUse %s <command> -h for more info\n", cmdName)
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
	{
		UsageLine: "version",
		Short:     "Print version information",
		Run: func(args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("no arguments allowed")
			}
			fmt.Println(version)
			return nil
		},
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
		cmd.Flag.Usage = func() { cmd.Usage() }
		cmd.Flag.Parse(args[1:])
		args = cmd.Flag.Args()
		if err := cmd.Run(args); err != nil {
			errExit("%s: %v\n", cmdName, err)
		}
		return
	}

	switch cmdName {
	default:
		fmt.Fprintf(os.Stderr, "Unrecognised command '%s'\n\n", cmdName)
		flag.Usage()
		os.Exit(2)
	}
}
