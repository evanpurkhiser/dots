package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go.evanpurkhiser.com/dots/config"
)

var (
	sourceConfig   *config.SourceConfig
	sourceLockfile *config.SourceLockfile
)

func main() {
	path := config.SourceConfigPath()

	loadConfigs := func(cmd *cobra.Command, args []string) error {
		var err error

		sourceConfig, err = config.LoadSourceConfig(path)
		if err != nil {
			return err
		}

		sourceLockfile, err = config.LoadLockfile(sourceConfig)
		if err != nil {
			return err
		}

		errs := config.SanitizeSourceConfig(sourceConfig)
		if len(errs) > 0 {
			fmt.Println(errs)
		}

		// TODO: If sanitization fails ask if we should continue anyway on destructive actions
		// Maybe ask to show dry run with verbose output before proceeding?

		return nil
	}

	cobra.EnableCommandSorting = false

	rootCmd := cobra.Command{
		Use:   "dots",
		Short: "A portable tool for managing a single set of dotfiles",

		SilenceUsage:      true,
		SilenceErrors:     true,
		PersistentPreRunE: loadConfigs,
	}

	rootCmd.AddCommand(&filesCmd)
	rootCmd.AddCommand(&diffCmd)
	rootCmd.AddCommand(&configCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//	fmt.Println(lockfile)
	//
	//	configs := resolver.ResolveConfigurations(resolver.Config{
	//		Groups:         source.Groups,
	//		SourcePath:     source.SourcePath,
	//		OverrideSuffix: source.OverrideSuffix,
	//	})
	//
	//	fmt.Println(configs["vim/config.vim"])

	// TODO: Output writer that looks somthing like
	//
	// source:  /home/.local/etc
	// install: /home/.config
	//
	// [=> base]     -- bash/bashrc
	// [=> composed] -- bash/environment
	//  -> composing from base and machines/desktop groups
	// [=> removed]  -- bash/bad-filemulti
	// [=> compiled] -- bash/complex
	//  -> ignoring configs in base and common/work due to override
	//  -> override file present in common/work-vm
	//  -> composing from machines/crunchydev (spliced at common/work-vm:22)
	//
	// [=> install script] base/bash.install
	//  -> triggered by base/bash/bashrc
	//  -> triggered by base/bash/environment

	// CLI Interfaceo

	// dots {config, install, diff, files, help}

	// dots install [filter...]
	// dots diff    [filter...]
	// dots files   [filter...]

	// dots config  {profiles, groups, use, override}

}
