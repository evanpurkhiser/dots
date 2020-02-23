package main

import (
	"os"
	"runtime/debug"

	"github.com/fatih/color"
	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"

	"go.evanpurkhiser.com/dots/config"
)

var Version = "dev"

var (
	sourceConfig   *config.SourceConfig
	sourceLockfile *config.SourceLockfile
)

func loadConfigs(cmd *cobra.Command, args []string) error {
	var err error

	path := config.SourceConfigPath()

	sourceConfig, err = config.LoadSourceConfig(path)
	if err != nil {
		return err
	}

	sourceLockfile, err = config.LoadLockfile(sourceConfig)
	if err != nil {
		return err
	}

	warns := config.SanitizeSourceConfig(sourceConfig)
	for _, err := range warns {
		color.New(color.FgYellow).Fprintf(os.Stderr, "warn: %s\n", err)
	}

	return nil
}

func sentryRecover() {
	err := recover()
	if err == nil {
		return
	}

	debug.PrintStack()
	sentry.CurrentHub().Recover(err)
}

func main() {
	sentry.Init(sentry.ClientOptions{
		Dsn:       "https://4c3f2bfcecf64bda8a4729f205e9a540@sentry.io/1522580",
		Transport: sentry.NewHTTPSyncTransport(),
	})

	defer sentryRecover()

	cobra.EnableCommandSorting = false

	rootCmd := cobra.Command{
		Use:     "dots",
		Short:   "A portable tool for managing a single set of dotfiles",
		Version: Version,

		SilenceUsage:      true,
		SilenceErrors:     true,
		PersistentPreRunE: loadConfigs,
	}

	rootCmd.AddCommand(&filesCmd)
	rootCmd.AddCommand(&diffCmd)
	rootCmd.AddCommand(&installCmd)
	rootCmd.AddCommand(&configCmd)

	if err := rootCmd.Execute(); err != nil {
		color.New(color.FgRed).Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
