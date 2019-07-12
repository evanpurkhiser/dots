package installer

import (
	"fmt"
	"io"
	"os"
	"path"
	"sync"

	"go.evanpurkhiser.com/dots/config"
	"go.evanpurkhiser.com/dots/resolver"
)

const separator = string(os.PathSeparator)

// directoryMode is the mode used to create directories for installed dotfiles.
const directoryMode = 0755

// InstallConfig represents configuration options available for installing
// a single or set of dotfiles.
type InstallConfig struct {
	SourceConfig *config.SourceConfig

	// OverrideInstallPath specifies a path to install the dotfile at,
	// overriding the configuration in the SourceConfig.
	OverrideInstallPath string

	// ForceReinstall installs the dotfile even if the dotfile has not been
	// changed from its source. This implies that install scripts will be run.
	ForceReinstall bool

	// SkipInstallScripts disables execution of any triggered install scripts
	SkipInstallScripts bool

	// TODO We can probably add a channel here to pipe logging output so that we
	// can output some logging
}

// InstalledDotfile is a represents of the dotfile *after* it has been
// installed into the configuration directory.
type InstalledDotfile struct {
	*PreparedDotfile

	// InstallError represents an error that occurred during installation.
	InstallError error
}

// InstallDotfile is given a prepared dotfile and installation configuration
// and will perform all the necessary actions to install the file into it's
// target location.
func InstallDotfile(dotfile *PreparedDotfile, config InstallConfig) error {
	installPath := config.SourceConfig.InstallPath + separator + dotfile.Path

	if config.OverrideInstallPath != "" {
		installPath = config.OverrideInstallPath + separator + dotfile.Path
	}

	if dotfile.SourcesAreIrregular {
		return fmt.Errorf("Source files are not all regular files")
	}

	if !dotfile.IsChanged() && !config.ForceReinstall {
		return nil
	}

	// Removed
	if dotfile.Removed && !dotfile.RemovedNull {
		return os.Remove(installPath)
	}

	targetMode := dotfile.Permissions.New | dotfile.Mode.New

	// Only filemode differs
	modeChanged := !dotfile.IsNew &&
		!dotfile.ContentsDiffer &&
		(dotfile.Permissions.IsChanged() || dotfile.Mode.IsChanged())

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

// RunInstallScripts executes all install scripts for a single dotfile.
func RunInstallScripts(dotfile resolver.Dotfile, config InstallConfig) error {
	// TODO actually implement this
	fmt.Println(dotfile.InstallScripts)

	return nil
}

// InstalledDotfiles asynchronously calls InstalledDotfile on all passed
// PreparedDotfiles. Once all dotfiles have been installed, all install scripts
// will execute, in order of installation (unless SkipInstallScripts is on).
func InstallDotfiles(dotfiles PreparedDotfiles, config InstallConfig) []*InstalledDotfile {
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(dotfiles))

	installed := make([]*InstalledDotfile, len(dotfiles))

	for i, dotfile := range dotfiles {
		go func(dotfile *PreparedDotfile) {
			err := InstallDotfile(dotfile, config)
			installed[i] = &InstalledDotfile{
				PreparedDotfile: dotfile,
				InstallError:    err,
			}
			waitGroup.Done()
		}(dotfile)
	}

	waitGroup.Wait()

	// Nothing left to do if there are no install scripts to run
	if config.SkipInstallScripts {
		return installed
	}

	// TODO After all dotfiles are installed, we now must run our installation scripts

	return installed
}
