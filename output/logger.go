package output

import (
	"fmt"
	"strings"

	"github.com/fatih/color"

	"go.evanpurkhiser.com/dots/config"
	"go.evanpurkhiser.com/dots/events"
	"go.evanpurkhiser.com/dots/installer"
)

var e = color.RedString("errn:")
var w = color.YellowString("warn:")
var n = color.CyanString("note:")

// Config is a object used to configure the output logger.
type Config struct {
	SourceConfig    config.SourceConfig
	InstallConfig   installer.InstallConfig
	PreparedInstall installer.PreparedInstall
	IsVerbose       bool
}

// New creates a output logger given a configuration.
func New(config Config) *Output {
	logger := &Output{
		Config:    config,
		eventChan: make(chan events.Event),
	}

	// Get the max length of the groups
	maxDotfileLength := 0
	for _, d := range config.PreparedInstall.Dotfiles {
		if logger.shouldLogDotfile(d) && len(d.Path) > maxDotfileLength {
			maxDotfileLength = len(d.Path)
		}
	}
	logger.maxDotfileLength = maxDotfileLength

	return logger
}

// Output represents a service object used to output logging information about
// dotfile installation operations.
type Output struct {
	Config
	eventChan        chan events.Event
	maxDotfileLength int
}

// shouldLogDotfile indicates if the dotfile should be logged given the current
// Output configuration.
func (l *Output) shouldLogDotfile(dotfile *installer.PreparedDotfile) bool {
	return dotfile.PrepareError != nil || installer.WillInstallDotfile(dotfile, l.InstallConfig)
}

func (l *Output) logEvent(event events.Event) {
	// TODO: Do something here
}

// GetEventChan returns the event channel that may be sent events to be
// outputted while the logger listening for events.
func (l *Output) GetEventChan() chan<- events.Event {
	return l.eventChan
}

// LogEvents processes the Events channel and output logging for each event
// processed on the channel. A function is returned that when called will stop
// processing.
//
// Typically this should be called with a defer, e.g:
//
//   defer myLogger.LogEvents()()
//
func (l *Output) LogEvents() func() {
	stop := make(chan bool)

	go func() {
		for {
			select {
			case <-stop:
				return
			case event := <-l.eventChan:
				l.logEvent(event)
			}
		}
	}()

	return func() { stop <- true }
}

// DryrunInstall outputs the logging of a dryrun of the prepared dotfiles
func (l *Output) DryrunInstall() {
	fmt.Printf("%s %s\n\n", n, "dry run — no dotfiles will be installed")
	for _, dotfile := range l.PreparedInstall.Dotfiles {
		l.DotfileInfo(dotfile)
	}
}

// InstallInfo outputs details about the pending installation. Output is only
// performed when verbosity is enabled.
func (l *Output) InstallInfo() {
	if !l.IsVerbose {
		return
	}

	fmt.Printf(
		"%s %s added %s removed %s modified %s error\n",
		color.HiBlackString("legend:"),
		color.HiGreenString("◼️"),
		color.HiRedString("◼️"),
		color.HiBlueString("◼️"),
		color.HiRedString("⨉"),
	)

	fmt.Printf("%s %s\n", color.HiBlackString("source:"), l.SourceConfig.SourcePath)
	fmt.Printf("%s %s\n", color.HiBlackString("target:"), l.SourceConfig.InstallPath)
	fmt.Println()
}

// DotfileInfo outputs information about a single prepared dotfile. Will not
// output anything without IsInfo. When IsVerbose is enabled additional
// information about the prepared dotfile will be included.
func (l *Output) DotfileInfo(dotfile *installer.PreparedDotfile) {
	if !l.IsVerbose {
		return
	}

	if !l.shouldLogDotfile(dotfile) {
		return
	}

	indicator := "◼️"
	indicatorColor := color.New()

	switch {
	case dotfile.PrepareError != nil:
		indicator = "⨉"
		indicatorColor.Add(color.FgRed)
	case dotfile.IsNew:
		indicatorColor.Add(color.FgHiGreen)
	case dotfile.Removed:
		indicatorColor.Add(color.FgHiRed)
	case dotfile.IsChanged():
		indicatorColor.Add(color.FgBlue)
	default:
		indicator = "-"
		indicatorColor.Add(color.FgHiBlack)
	}

	group := ""
	if len(dotfile.Sources) == 1 {
		group = dotfile.Sources[0].Group
	} else {
		groups := make([]string, 0, len(dotfile.Sources))

		for _, source := range dotfile.Sources {
			groups = append(groups, source.Group)
		}

		group = strings.Join(groups, " ")
	}

	group = fmt.Sprintf(
		"%s %s %s",
		color.HiBlackString("["),
		color.HiWhiteString(group),
		color.HiBlackString("]"),
	)

	output := fmt.Sprintf(" %%s %%-%ds %%s", l.maxDotfileLength+1)
	fmt.Printf(
		output,
		indicatorColor.Sprint(indicator),
		dotfile.Path,
		group,
	)

	if dotfile.Permissions.IsChanged() {
		fmt.Printf(
			" [%s %s %s]",
			color.New(color.FgHiRed).Sprintf("%#o", int(dotfile.Permissions.Old)),
			color.HiWhiteString("→"),
			color.New(color.FgHiGreen).Sprintf("%#o", int(dotfile.Permissions.New)),
		)
	}

	fmt.Println()

	ln := func(v ...interface{}) {
		fmt.Printf("   %s %s\n", v...)
	}

	if dotfile.PrepareError != nil {
		ln(e, color.HiRedString(dotfile.PrepareError.Error()))
	}

	if dotfile.OverwritesExisting {
		ln(w, "overwriting existing file")
	}

	if dotfile.SourcePermissionsDiffer {
		ln(w, "inconsistent source file permissions")
	}

	if dotfile.RemovedNull {
		ln(n, "nothing to remove")
	}
}
