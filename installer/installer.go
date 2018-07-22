package installer

import (
	"fmt"
	"io"
	"os"
	"path"
	"sync"

	"go.evanpurkhiser.com/dots/config"
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
}

func InstallDotfile(dotfile *PreparedDotfile, config InstallConfig) error {
	installPath := config.SourceConfig.InstallPath + separator + dotfile.Path

	if config.OverrideInstallPath != "" {
		installPath = config.OverrideInstallPath + separator + dotfile.Path
	}

	if dotfile.SourcesAreIrregular {
		return fmt.Errorf("Source files are not all regular files")
	}

	// No change
	if !dotfile.IsNew && !dotfile.ContentsDiffer && dotfile.Permissions.IsSame() {
		return nil
	}

	// Removed
	if dotfile.Removed && !dotfile.RemovedNull {
		return os.Remove(installPath)
	}

	targetMode := dotfile.Permissions.New | dotfile.Mode

	// Only permissions differ
	if !dotfile.IsNew && !dotfile.Permissions.IsSame() && !dotfile.ContentsDiffer {
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

func InstallDotfiles(dotfiles PreparedDotfiles, config InstallConfig) []error {
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(dotfiles))

	errors := []error{}

	for _, dotfile := range dotfiles {
		go func(dotfile *PreparedDotfile) {
			err := InstallDotfile(dotfile, config)
			if err != nil {
				errors = append(errors, err)
			}
			waitGroup.Done()
		}(dotfile)
	}

	waitGroup.Wait()

	return errors
}
