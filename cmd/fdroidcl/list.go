// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"
	"os"
	"sort"
)

var cmdList = &Command{
	UsageLine: "list (categories)",
	Short:     "List all known values of a kind",
}

func init() {
	cmdList.Run = runList
}

func runList(args []string) {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "need exactly one argument")
		cmdList.Flag.Usage()
	}
	apps := mustLoadIndexes()
	values := make(map[string]struct{})
	switch args[0] {
	case "categories":
		for _, app := range apps {
			for _, c := range app.Categs {
				values[c] = struct{}{}
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "invalid argument")
		cmdList.Flag.Usage()
	}
	result := make([]string, 0, len(values))
	for s := range values {
		result = append(result, s)
	}
	sort.Strings(result)
	for _, s := range result {
		fmt.Println(s)
	}
}
