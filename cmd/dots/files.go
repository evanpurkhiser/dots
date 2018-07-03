package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"go.evanpurkhiser.com/dots/resolver"
)

var filesCmd = cobra.Command{
	Use:   "files [filter...]",
	Short: "List resolved dotfile paths",
	RunE: func(cmd *cobra.Command, args []string) error {
		dotfiles := resolver.ResolveDotfiles(*sourceConfig, *sourceLockfile)

		fmt.Println(strings.Join(dotfiles.Filter(args).Files(), "\n"))

		return nil
	},
}
