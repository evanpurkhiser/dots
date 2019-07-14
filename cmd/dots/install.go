package main

import (
	"github.com/spf13/cobra"

	"go.evanpurkhiser.com/dots/installer"
	"go.evanpurkhiser.com/dots/resolver"
)

var installCmd = cobra.Command{
	Use:   "install [filter...]",
	Short: "Install and compile dotfiles from sources",
	RunE: func(cmd *cobra.Command, args []string) error {
		forceReInstall, _ := cmd.Flags().GetBool("reinstall")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		dotfiles := resolver.ResolveDotfiles(*sourceConfig, *sourceLockfile).Filter(args)
		prepared := installer.PrepareDotfiles(dotfiles, *sourceConfig)

		if dryRun {
			// TODO: Do logging output
			return nil
		}

		config := installer.InstallConfig{
			SourceConfig:   sourceConfig,
			ForceReinstall: forceReInstall,
		}

		installer.InstallDotfiles(prepared, config)
		installer.RunInstallScripts(prepared, config)

		return nil
	},
	Args: cobra.ArbitraryArgs,
}

func init() {
	flags := installCmd.Flags()
	flags.SortFlags = false

	flags.BoolP("reinstall", "r", false, "forces execution of all installation scripts")
	flags.BoolP("dry-run", "n", false, "do not mutate any dotfiles, implies info")
	flags.BoolP("info", "i", false, "prints install operation details")
	flags.BoolP("verbose", "v", false, "prints debug data, implies info")
}
