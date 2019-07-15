package installer

import (
	"fmt"
	"os"
	"os/exec"
)

// RunInstallScripts executes the installation scripts of a PreparedInstall for
// files that have changed (unless ForceReinstall is enabled, in which case
// *all* scripts will be run).
func RunInstallScripts(install PreparedInstall, config InstallConfig) {
	for _, script := range install.InstallScripts {
		if !script.ShouldInstall() && !config.ForceReinstall {
			continue
		}

		if !script.Executable {
			continue
		}

		command := exec.Command(script.Path)

		// Execute the script in the installed path cotnext
		if config.OverrideInstallPath != "" {
			command.Dir = config.OverrideInstallPath
		} else {
			command.Dir = config.SourceConfig.InstallPath
		}

		// Setup the environment
		command.Env = append(
			os.Environ(),
			fmt.Sprintf("DOTS_SOURCE=%s", config.SourceConfig.SourcePath),
			fmt.Sprintf("DOTS_FORCE_REINSTALL=%t", config.ForceReinstall),
		)

		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		command.Run()
	}

}
