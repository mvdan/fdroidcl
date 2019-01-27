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

func runList(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("need exactly one argument")
	}
	apps, err := loadIndexes()
	if err != nil {
		return err
	}
	values := make(map[string]struct{})
	switch args[0] {
	case "categories":
		for _, app := range apps {
			for _, c := range app.Categories {
				values[c] = struct{}{}
			}
		}
	default:
		return fmt.Errorf("invalid argument")
	}
	result := make([]string, 0, len(values))
	for s := range values {
		result = append(result, s)
	}
	sort.Strings(result)
	for _, s := range result {
		fmt.Fprintln(os.Stdout, s)
	}
	return nil
}
