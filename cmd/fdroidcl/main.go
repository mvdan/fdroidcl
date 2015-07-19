// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mvdan/basedir"
)

const (
	cmdName = "fdroidcl"

	repoName = "repo"
	repoURL  = "https://f-droid.org/repo"
)

func subdir(dir, name string) string {
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(p, 0755); err != nil {
		log.Fatalf("Could not create dir '%s': %v", p, err)
	}
	return p
}

func mustCache() string {
	dir, err := basedir.Cache()
	if err != nil {
		log.Fatalf("Could not determine cache dir: %v", err)
	}
	return subdir(dir, cmdName)
}

func mustConfig() string {
	dir, err := basedir.Config()
	if err != nil {
		log.Fatalf("Could not determine config dir: %v", err)
	}
	return subdir(dir, cmdName)
}

type Repo struct {
	URL string `json:"url"`
}

type userConfig struct {
	Repos map[string]Repo `json:"repos"`
}

var config = userConfig{
	Repos: map[string]Repo{
		"f-droid": {
			URL: "https://f-droid.org/repo",
		},
	},
}

func readConfig() {
	p := filepath.Join(mustConfig(), "config.json")
	f, err := os.Open(p)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("Error when opening config file: %v", err)
		}
		f, err := os.Create(p)
		if err != nil {
			log.Fatalf("Error when creating config file: %v", err)
		}
		defer f.Close()
		b, err := json.MarshalIndent(&config, "", "\t")
		if err != nil {
			log.Fatalf("Error when encoding config file: %v", err)
		}
		if _, err := f.Write(b); err != nil {
			log.Fatalf("Error when writing config file: %v", err)
		}
		return
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		log.Fatalf("Error when decoding config file: %v", err)
	}
}

// A Command is an implementation of a go command
// like go build or go fix.
type Command struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(args []string)

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
		fmt.Fprintf(os.Stderr, "\nUse %s <command> -h for more info\n", cmdName)
	}
}

// Commands lists the available commands.
var commands = []*Command{
	cmdUpdate,
	cmdSearch,
	cmdShow,
	cmdDevices,
	cmdDownload,
	cmdInstall,
	cmdUninstall,
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		flag.Usage()
		os.Exit(2)
	}

	for _, cmd := range commands {
		if cmd.Name() != args[0] {
			continue
		}
		readConfig()
		cmd.Flag.Usage = func() { cmd.Usage() }
		cmd.Flag.Parse(args[1:])
		args = cmd.Flag.Args()
		cmd.Run(args)
		os.Exit(0)
	}

	switch args[0] {
	default:
		log.Printf("Unrecognised command '%s'\n\n", args[0])
		flag.Usage()
		os.Exit(2)
	}
}
