package installer

import (
	"io"
	"os"
	"path"
	"sync"

	"go.evanpurkhiser.com/dots/config"
	"go.evanpurkhiser.com/dots/events"
)

const separator = string(os.PathSeparator)

// directoryMode is the mode used to create directories for installed dotfiles.
const directoryMode = 0755

// InstallConfig represents configuration options available for installing
// a single or set of dotfiles.
type InstallConfig struct {
	SourceConfig   *config.SourceConfig
	SourceLockfile *config.SourceLockfile

	// OverrideInstallPath specifies a path to install the dotfile at,
	// overriding the configuration in the SourceConfig.
	OverrideInstallPath string

	// ForceReinstall installs the dotfile even if the dotfile has not been
	// changed from its source. This implies that install scripts will be run.
	ForceReinstall bool

	// EventLogger may be given data to output during the installation process.
	// It is up to the logger service to decide if it can output anything for the
	// given data or not.
	EventLogger chan<- events.Event
}

// InstalledDotfile is a represents of the dotfile *after* it has been
// installed into the configuration directory.
type InstalledDotfile struct {
	*PreparedDotfile

	// InstallError represents an error that occurred during installation.
	InstallError error
}

// InstalledDotfiles represents a set of installed dotfiles.
type InstalledDotfiles []*InstalledDotfile

// HadError indicates if any dotfiles had errors while preparing or installing
func (i *InstalledDotfiles) HadError() bool {
	for _, dotfile := range *i {
		if dotfile.PrepareError != nil || dotfile.InstallError != nil {
			return true
		}
	}

	return false
}

// WillInstallDotfile indicates weather the dotfile will be installed if
// InstallDotfile is given the prepared dotfile. This does not guarentee that
// errors will not occur during installation.
func WillInstallDotfile(dotfile *PreparedDotfile, config InstallConfig) bool {
	// Skip dotfiles that we failed to prepare
	if dotfile.PrepareError != nil {
		return false
	}

	if !dotfile.IsChanged() && !config.ForceReinstall {
		return false
	}

	return true
}

// InstallDotfile is given a prepared dotfile and installation configuration
// and will perform all the necessary actions to install the file into it's
// target location.
func InstallDotfile(dotfile *PreparedDotfile, config InstallConfig) error {
	if !WillInstallDotfile(dotfile, config) {
		return nil
	}

	installPath := config.SourceConfig.InstallPath + separator + dotfile.Path

	if config.OverrideInstallPath != "" {
		installPath = config.OverrideInstallPath + separator + dotfile.Path
	}

	// Removed
	if dotfile.Removed && !dotfile.RemovedNull {
		return os.Remove(installPath)
	}

	targetMode := dotfile.Permissions.New

	// Only filemode differs
	modeChanged := !dotfile.IsNew &&
		!dotfile.ContentsDiffer && dotfile.Permissions.IsChanged()

	if modeChanged {
		return os.Chmod(installPath, targetMode)
	}

	if err := os.MkdirAll(path.Dir(installPath), directoryMode); err != nil {
		return err
	}

	targetOpts := os.O_CREATE | os.O_TRUNC | os.O_WRONLY

	target, err := os.OpenFile(installPath, targetOpts, targetMode)
	if err != nil {
		return err
	}
	defer target.Close()

	source, err := OpenDotfile(dotfile.Dotfile, *config.SourceConfig)
	if err != nil {
		return err
	}
	defer source.Close()

	_, err = io.Copy(target, source)

	return err
}

// InstallDotfiles asynchronously calls InstalledDotfile on all passed
// PreparedDotfiles.
func InstallDotfiles(install PreparedInstall, config InstallConfig) InstalledDotfiles {
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(install.Dotfiles))

	installed := make(InstalledDotfiles, len(install.Dotfiles))

	doInstall := func(i int, dotfile *PreparedDotfile) {
		err := InstallDotfile(dotfile, config)

		installed[i] = &InstalledDotfile{
			PreparedDotfile: dotfile,
			InstallError:    err,
		}

		config.EventLogger <- events.Event{
			Type:   events.DotfileInstalled,
			Object: installed[i],
		}

		waitGroup.Done()
	}

	config.EventLogger <- events.Event{
		Type:   events.InstallStarting,
		Object: install,
	}

	for i, dotfile := range install.Dotfiles {
		go doInstall(i, dotfile)
	}

	waitGroup.Wait()

	config.EventLogger <- events.Event{
		Type:   events.InstallDone,
		Object: installed,
	}

	return installed
}

// FinalizeInstall writes the updated lockfile after installation
func FinalizeInstall(installed []*InstalledDotfile, installConfig InstallConfig) error {
	installedFiles := make([]string, 0, len(installed))

	for _, dotfile := range installed {
		if dotfile.Removed {
			continue
		}
		if dotfile.InstallError != nil {
			continue
		}

		installedFiles = append(installedFiles, dotfile.Path)
	}

	lockfile := installConfig.SourceLockfile
	lockfile.InstalledFiles = installedFiles

	return config.WriteLockfile(lockfile, installConfig.SourceConfig)
}
