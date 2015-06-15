/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	repoName = "repo"
	repoURL  = "https://f-droid.org/repo"
)

// A Command is an implementation of a go command
// like go build or go fix.
type Command struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(args []string)

	// Name of the command.
	Name string

	// Short is the short description.
	Short string

	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet
}

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: fdroidcl [-h] <command> [<args>]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Available commands:")
		fmt.Fprintln(os.Stderr, "   update             Update the index")
		fmt.Fprintln(os.Stderr, "   list               List all available apps")
		fmt.Fprintln(os.Stderr, "   search <term...>   Search available apps")
		fmt.Fprintln(os.Stderr, "   show <appid...>    Show detailed info of an app")
		fmt.Fprintln(os.Stderr, "   devices            List connected devices")
		fmt.Fprintln(os.Stderr, "   installed          List installed apps")
	}
}

// Commands lists the available commands.
var commands = []*Command{
	cmdUpdate,
	cmdList,
	cmdSearch,
	cmdShow,
	cmdDevices,
	cmdInstalled,
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		flag.Usage()
		os.Exit(2)
	}

	for _, cmd := range commands {
		if cmd.Name != args[0] {
			continue
		}
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
