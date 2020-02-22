package events

type EventType string

// Available event types.
const (
	// InstallStarting fires when an installation for a set of dotfiles has bugun
	// processing.
	InstallStarting EventType = "install_starting"

	// InstallDone fires when all dotfiles have been installed. No scripts
	// have been executed for the dotfiles yet.
	InstallDone EventType = "install_completed"

	// DotfileInstalled fires when a single dotfile has completed installation.
	// This does not indicate that the insatllation was not a no-op.
	DotfileInstalled EventType = "dotfile_installed"

	// ScriptExecStarted fires when script execution has begun.
	ScriptExecStarted EventType = "executing_all_scripts"

	// ScriptExecDone fires when script execution has begun.
	ScriptExecDone EventType = "executing_all_scripts"

	// ScriptExecuting fires when a dotfile script is being executed.
	ScriptExecuting EventType = "executing_script"

	// ScriptCompleted fires when a dotfiles script has completed execution.
	ScriptCompleted EventType = "script_completed"
)

// Event represents a output logging event.
type Event struct {
	Type   EventType
	Object interface{}
}
