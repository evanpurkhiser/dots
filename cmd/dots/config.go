package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"go.evanpurkhiser.com/dots/config"
)

var configCmd = cobra.Command{
	Use:   "config",
	Short: "Manage configuration of profiles and groups",
}

var configGroupsCmd = cobra.Command{
	Use:   "groups",
	Short: "List all configured source groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(strings.Join(sourceConfig.Groups, "\n"))

		return nil
	},
	Args: cobra.NoArgs,
}

var configProfilesCmd = cobra.Command{
	Use:   "profiles",
	Short: "List all configured source profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(strings.Join(sourceConfig.Profiles.Names(), "\n"))

		return nil
	},
	Args: cobra.NoArgs,
}

var configActiveCmd = cobra.Command{
	Use:   "active",
	Short: "Shows the current active profile and groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		color.New(color.FgHiBlack).Print("profile: ")
		if sourceLockfile.Profile == "" {
			fmt.Println("<no profile>")
		} else {
			fmt.Println(sourceLockfile.Profile)
		}

		color.New(color.FgHiBlack).Print("groups:  ")
		if groups, ok := sourceConfig.Profiles[sourceLockfile.Profile]; ok {
			fmt.Printf("[%s]\n", strings.Join(groups, ", "))
		} else {
			fmt.Printf("[%s]\n", strings.Join(sourceLockfile.Groups, ", "))
		}

		return nil
	},
	Args: cobra.NoArgs,
}

var configUseCmd = cobra.Command{
	Use:   "use [profile]",
	Short: "Enable a specific profile for this host",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}

		sourceLockfile.Profile = args[0]
		sourceLockfile.Groups = []string{}

		err := config.ValidateLockfile(sourceLockfile, sourceConfig)
		if err != nil {
			return err
		}

		return config.WriteLockfile(sourceLockfile, sourceConfig)
	},
}

var configOverrideCmd = cobra.Command{
	Use:   "override [groups...]",
	Short: "Configure the host with ad hoc groups instead of a profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Usage()
		}

		sourceLockfile.Profile = ""
		sourceLockfile.Groups = args

		err := config.ValidateLockfile(sourceLockfile, sourceConfig)
		if err != nil {
			return err
		}

		return config.WriteLockfile(sourceLockfile, sourceConfig)
	},
}

var configClearCmd = cobra.Command{
	Use:   "clear",
	Short: "Clears profile and group configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceLockfile.Profile = ""
		sourceLockfile.Groups = []string{}

		return config.WriteLockfile(sourceLockfile, sourceConfig)
	},
}

func init() {
	configCmd.AddCommand(&configUseCmd)
	configCmd.AddCommand(&configOverrideCmd)
	configCmd.AddCommand(&configProfilesCmd)
	configCmd.AddCommand(&configGroupsCmd)
	configCmd.AddCommand(&configActiveCmd)
	configCmd.AddCommand(&configClearCmd)
}
