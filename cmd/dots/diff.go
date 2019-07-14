package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"go.evanpurkhiser.com/dots/installer"
	"go.evanpurkhiser.com/dots/resolver"

	"github.com/spf13/cobra"
)

var diffCmd = cobra.Command{
	Use:   "diff [git-diff options] [filter...]",
	Short: "Compare the currently installed dotfiles to their sources",
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
			return fmt.Errorf("failed to create tmp directory: %s", err)
		}
		defer os.RemoveAll(sourceTmp)

		// Create a "pretty link" so that our diff looks nicer
		prettyLink := sourceConfig.InstallPath + "-staged"
		if err := os.Symlink(sourceTmp, prettyLink); err != nil {
			return fmt.Errorf("failed to create tmp symlink: %s", err)
		}
		defer os.Remove(prettyLink)

		installConfig := installer.InstallConfig{
			SourceConfig:        sourceConfig,
			OverrideInstallPath: sourceTmp,
		}

		dotfiles := resolver.ResolveDotfiles(*sourceConfig, *sourceLockfile)
		prepared := installer.PrepareDotfiles(dotfiles.Filter(files), *sourceConfig)
		installer.InstallDotfiles(prepared, installConfig)

		git := []string{"diff", "--no-index", "--diff-filter=MA"}
		git = append(git, flags...)
		git = append(git, "--", sourceConfig.InstallPath, prettyLink+"/")

		command := exec.Command("git", git...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		command.Run()

		return nil
	},

	Args:                  cobra.ArbitraryArgs,
	DisableFlagsInUseLine: true,
	DisableFlagParsing:    true,
}
