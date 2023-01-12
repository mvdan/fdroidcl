// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"fmt"
)

var cmdRepo = &Command{
	UsageLine: "repo",
	Short:     "Manage repositories",
	Long: `
List, add, remove, enable or disable repositories.
When a repository is added, it is enabled by default.

List repositories:

	$ fdroidcl repo

Modify repositories:

	$ fdroidcl repo add <NAME> <URL>
	$ fdroidcl repo remove <NAME>
	$ fdroidcl repo enable <NAME>
	$ fdroidcl repo disable <NAME>
`[1:],
}

func init() {
	cmdRepo.Run = runRepo
}

func runRepo(args []string) error {
	if len(args) == 0 {
		// list repositories
		for i, value := range config.Repos {
			fmt.Printf("Name: %s\n", value.ID)
			fmt.Printf("URL: %s\n", value.URL)
			var enabled string
			if value.Enabled {
				enabled = "yes"
			} else {
				enabled = "no"
			}
			fmt.Printf("Enabled: %s\n", enabled)
			if i != len(config.Repos)-1 {
				fmt.Println()
			}
		}
		return nil
	}
	if args[0] == "add" {
		if len(args) != 3 {
			return fmt.Errorf("wrong amount of arguments")
		}
		return addRepo(args[1], args[2])
	} else if args[0] == "remove" {
		if len(args) != 2 {
			return fmt.Errorf("wrong amount of arguments")
		}
		return removeRepo(args[1])
	} else if args[0] == "enable" {
		if len(args) != 2 {
			return fmt.Errorf("wrong amount of arguments")
		}
		return enableRepo(args[1])
	} else if args[0] == "disable" {
		if len(args) != 2 {
			return fmt.Errorf("wrong amount of arguments")
		}
		return disableRepo(args[1])
	} else {
		return fmt.Errorf("wrong usage")
	}
}

func repoIndex(name string) int {
	index := -1
	for i, value := range config.Repos {
		if value.ID == name {
			index = i
			break
		}
	}
	return index
}

func addRepo(name, url string) error {
	if repoIndex(name) != -1 {
		return fmt.Errorf("a repo with the same name \"%s\" exists already", name)
	}
	config.Repos = append(config.Repos, repo{ID: name, URL: url, Enabled: true})
	return writeConfig(&config)
}

func removeRepo(name string) error {
	index := repoIndex(name)
	if index == -1 {
		return fmt.Errorf("a repo with the name \"%s\" could not be found", name)
	}
	config.Repos = append(config.Repos[:index], config.Repos[index+1:]...)
	return writeConfig(&config)
}

func enableRepo(name string) error {
	index := repoIndex(name)
	if index == -1 {
		return fmt.Errorf("a repo with the name \"%s\" could not be found", name)
	}
	config.Repos[index].Enabled = true
	return writeConfig(&config)
}

func disableRepo(name string) error {
	index := repoIndex(name)
	if index == -1 {
		return fmt.Errorf("a repo with the name \"%s\" could not be found", name)
	}
	config.Repos[index].Enabled = false
	return writeConfig(&config)
}
