package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.evanpurkhiser.com/dots/installer"
	"go.evanpurkhiser.com/dots/output"
	"go.evanpurkhiser.com/dots/resolver"
)

var installCmd = cobra.Command{
	Use:   "install [filter...]",
	Short: "Install and compile dotfiles from sources",
	RunE: func(cmd *cobra.Command, args []string) error {
		forceReInstall, _ := cmd.Flags().GetBool("reinstall")
		verbose, _ := cmd.Flags().GetBool("verbose")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			verbose = true
		}

		dotfiles := resolver.ResolveDotfiles(*sourceConfig, *sourceLockfile).Filter(args)
		prepared := installer.PrepareDotfiles(dotfiles, *sourceConfig)

		installConfig := installer.InstallConfig{
			SourceConfig:   sourceConfig,
			SourceLockfile: sourceLockfile,
			ForceReinstall: forceReInstall,
		}

		installLogger := output.New(output.Config{
			SourceConfig:    *sourceConfig,
			InstallConfig:   installConfig,
			PreparedInstall: prepared,
			IsVerbose:       verbose,
		})

		installConfig.EventLogger = installLogger.GetEventChan()

		installLogger.InstallInfo()

		if dryRun {
			installLogger.DryrunInstall()
			return nil
		}

		defer installLogger.LogEvents()()

		installed := installer.InstallDotfiles(prepared, installConfig)
		executedScripts := installer.RunInstallScripts(prepared, installConfig)
		finalizeErr := installer.FinalizeInstall(installed, installConfig)

		if installed.HadError() {
			return fmt.Errorf("some dotfiles failed to install")
		}

		if executedScripts.HadError() {
			return fmt.Errorf("some dotfiles scripts had errors")
		}

		if finalizeErr != nil {
			return fmt.Errorf("finalization error: %s", finalizeErr)
		}

		return nil
	},
	Args: cobra.ArbitraryArgs,
}

func init() {
	flags := installCmd.Flags()
	flags.SortFlags = false

	flags.BoolP("reinstall", "r", false, "forces execution of all installation scripts")
	flags.BoolP("verbose", "v", false, "prints debug data")
	flags.BoolP("dry-run", "n", false, "do not mutate any dotfiles, implies verbose")
}
