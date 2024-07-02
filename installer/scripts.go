package installer

import (
	"fmt"
	"os"
	"os/exec"

	"go.evanpurkhiser.com/dots/events"
)

// ExecutedScript represents the completion of a script execution
type ExecutedScript struct {
	*InstallScript

	// ExecutionError represents any error that occurred during script execution.
	ExecutionError error
}

// ExecutedScripts represents a set of executed scripts.
type ExecutedScripts []*ExecutedScript

// HadError indicates if any dotfiles had errors while preparing or installing
func (i *ExecutedScripts) HadError() bool {
	for _, script := range *i {
		if script.PrepareError != nil || script.ExecutionError != nil {
			return true
		}
	}

	return false
}

// RunInstallScript executes a single InstallScript.
func RunInstallScript(script *InstallScript, config InstallConfig) error {
	if !script.ShouldInstall() && !config.ForceReinstall {
		return nil
	}

	if !script.Executable {
		return nil
	}

	command := exec.Command(script.FilePath)

	// Execute the script in the installed path context
	if config.OverrideInstallPath != "" {
		command.Dir = config.OverrideInstallPath
	} else {
		command.Dir = config.SourceConfig.InstallPath
	}

	command.Dir = command.Dir + separator + script.Path

	// Setup the environment
	command.Env = append(
		os.Environ(),
		fmt.Sprintf("DOTS_SOURCE=%s", config.SourceConfig.SourcePath),
		fmt.Sprintf("DOTS_FORCE_REINSTALL=%t", config.ForceReinstall),
	)

	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	return command.Run()
}

// RunInstallScripts executes the installation scripts of a PreparedInstall for
// files that have changed (unless ForceReinstall is enabled, in which case
// *all* scripts will be run).
func RunInstallScripts(install PreparedInstall, config InstallConfig) ExecutedScripts {
	executedScripts := make(ExecutedScripts, len(install.InstallScripts))

	config.EventLogger <- events.Event{
		Type:   events.ScriptExecStarted,
		Object: install,
	}

	for i, script := range install.InstallScripts {
		config.EventLogger <- events.Event{
			Type:   events.ScriptCompleted,
			Object: script,
		}

		err := RunInstallScript(script, config)

		executedScripts[i] = &ExecutedScript{
			InstallScript:  script,
			ExecutionError: err,
		}

		config.EventLogger <- events.Event{
			Type:   events.ScriptCompleted,
			Object: script,
		}
	}

	config.EventLogger <- events.Event{
		Type:   events.ScriptExecDone,
		Object: executedScripts,
	}

	return executedScripts
}
