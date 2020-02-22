package events

// EventType represents the key of an event that may be triggered
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

// NewNoopLogger creates a NoopLogger.
func NewNoopLogger() *NoopLogger {
	logger := &NoopLogger{
		eventChan: make(chan Event),
	}

	return logger
}

// NoopLogger may be used to handle event processing without doing
// anything with the events.
type NoopLogger struct {
	eventChan chan Event
}

// GetEventChan returns the event channel in which no events will be logged.
func (l *NoopLogger) GetEventChan() chan<- Event {
	return l.eventChan
}

// LogEvents processes the Event channel, but will do nothing with the events
func (l *NoopLogger) LogEvents() func() {
	stop := make(chan bool)

	go func() {
		for {
			select {
			case <-l.eventChan:
				continue
			case <-stop:
				return
			}
		}
	}()

	return func() { stop <- true }
}
