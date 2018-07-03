package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
)

var diffCmd = cobra.Command{
	Use:   "diff [git-diff options] [filter...]",
	Short: "Manage configuration of profiles and groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		files := []string{}
		flags := []string{}

		for _, arg := range args {
			if arg == "--" {
				continue
			}

			if strings.HasPrefix(arg, "-") {
				flags = append(flags, arg)
			} else {
				files = append(files, arg)
			}
		}

		sourceTmp, err := ioutil.TempDir("", "dots-source")
		if err != nil {
			return fmt.Errorf("Failed to create tmp directory: %s\n", err)
		}

		activeTmp, err := ioutil.TempDir("", "dots-active")
		if err != nil {
			return fmt.Errorf("Failed to create tmp directory: %s\n", err)
		}

		fmt.Println(sourceTmp, activeTmp)

		exec := []string{"git", "diff", "--no-index"}
		exec = append(exec, flags...)
		exec = append(exec, "--", sourceTmp, activeTmp)

		fmt.Println(exec)

		return nil
	},

	Args: cobra.ArbitraryArgs,
	DisableFlagsInUseLine: true,
	DisableFlagParsing:    true,
}
