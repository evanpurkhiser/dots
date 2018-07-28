package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.evanpurkhiser.com/dots/installer"
	"go.evanpurkhiser.com/dots/resolver"
)

var installCmd = cobra.Command{
	Use:   "install [filter...]",
	Short: "Install and compile dotfiles from sources",
	RunE: func(cmd *cobra.Command, args []string) error {
		dotfiles := resolver.ResolveDotfiles(*sourceConfig, *sourceLockfile).Filter(args)
		prepared := installer.PrepareDotfiles(dotfiles, *sourceConfig)

		fmt.Println(prepared)

		return nil
	},
	Args: cobra.ArbitraryArgs,
}

func init() {
	flags := installCmd.Flags()
	flags.SortFlags = false

	flags.BoolP("reinstall", "r", false, "reinstall all dotfiles, includes unchanged files")
	flags.BoolP("dry-run", "n", false, "do not mutate any dotfiles, implies info")
	flags.BoolP("info", "i", false, "prints install operation details")
	flags.BoolP("verbose", "v", false, "prints debug data, implies info")
}
