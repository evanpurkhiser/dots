package installer

import (
	"os"
	"sync"

	"go.evanpurkhiser.com/dots/config"
	"go.evanpurkhiser.com/dots/resolver"
)

const separator = string(os.PathSeparator)

// A PreparedDotfile represents a dotfile that has been "prepared" for
// installation by verifying it's contents against the existing dotfile, and
// checking various other flags that require knowledge of the existing dotfile.
type PreparedDotfile struct {
	*resolver.Dotfile

	// InstallPath is the full path to which the dotfile will be installed.
	InstallPath string

	// ContentsDiffer is a boolean flag representing that the compiled source
	// differs from the currently installed dotfile.
	ContentsDiffer bool

	// SourceModesDiffer indicates that a compiled dotfile (one with multiple
	// sources) does not have a consistent mode across all sources. In this
	// case the lowest mode will be used.
	SourceModesDiffer bool

	// ModeDiffers represents the change in modes between the compiled source
	// and the currently installed dotfile. Equal modes can be verified by
	// calling ModeDiff.IsSame.
	ModeDiff ModeDiff

	// RemovedNull is a warning flag indicating that the removed dotfile does
	// not exist in the install tree, though the dotfile is marked as removed.
	RemovedNull bool

	// OverwritesExisting is a warning flag that indicates that installing this
	// dotfile is overwriting a dotfile that was not part of the lockfile.
	OverwritesExisting bool

	// PrepareError keeps track of errors while preparing the dotfile. Should
	// this contain any errors, the PreparedDotfile is likely incomplete.
	PrepareError error
}

// ModeDiff represents a change in file mode.
type ModeDiff struct {
	Old os.FileMode
	New os.FileMode
}

// IsSame returns a boolean value indicating if the modes are equal.
func (d ModeDiff) IsSame() bool {
	return d.New == d.Old
}

// PreparedDotfiles is a list of prepared dotfiles.
type PreparedDotfiles []*PreparedDotfile

// PrepareDotfiles iterates all passed dotfiles and creates an associated
// PreparedDotfile, returning a list of all prepared dotfiles.
func PrepareDotfiles(dotfiles resolver.Dotfiles, config config.SourceConfig) PreparedDotfiles {
	preparedDotfiles := make(PreparedDotfiles, len(dotfiles))

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(dotfiles))

	prepare := func(index int, dotfile *resolver.Dotfile) {
		defer waitGroup.Done()

		installPath := config.InstallPath + separator + dotfile.Path

		prepared := PreparedDotfile{
			Dotfile:     dotfile,
			InstallPath: installPath,
		}
		preparedDotfiles[index] = &prepared

		targetStat, targetStatErr := os.Lstat(installPath)

		exists := !os.IsNotExist(targetStatErr)

		// If we're unable to stat our target installation file and the file
		// exists there's likely a permissions issue.
		if targetStatErr != nil && exists {
			prepared.PrepareError = targetStatErr
			return
		}

		// Nothing needs to be verified if the dotfile is simply being added
		if dotfile.Added && !exists {
			return
		}

		if dotfile.Added && exists {
			prepared.OverwritesExisting = true
		}

		if dotfile.Removed && !exists {
			prepared.RemovedNull = true
		}

		sourceInfo := make([]os.FileInfo, len(dotfile.Sources))

		for i, source := range dotfile.Sources {
			path := config.SourcePath + separator + source.Path

			info, err := os.Lstat(path)
			if err != nil {
				prepared.PrepareError = err
				return
			}
			sourceInfo[i] = info
		}

		sourceMode, tookLowest := flattenModes(sourceInfo)

		prepared.ModeDiff = ModeDiff{
			Old: targetStat.Mode(),
			New: sourceMode,
		}
		prepared.SourceModesDiffer = tookLowest

		// If we are dealing with a dotfile with a single source we can quickly
		// determine modification based on differing sizes, otherwise we will
		// have to compare the compiled sources to the installed file.
		if len(dotfile.Sources) == 1 && targetStat.Size() != sourceInfo[0].Size() {
			prepared.ContentsDiffer = true
			return
		}

		// Compare source and currently instlled dotfile
		source, err := OpenDotfile(dotfile, config)
		if err != nil {
			prepared.PrepareError = err
			return
		}
		defer source.Close()

		target, err := os.Open(installPath)
		if err != nil {
			prepared.PrepareError = err
			return
		}
		defer target.Close()

		filesAreSame, err := compareReaders(source, target)
		if err != nil {
			prepared.PrepareError = err
		}

		prepared.ContentsDiffer = !filesAreSame
	}

	for i, dotfile := range dotfiles {
		go prepare(i, dotfile)
	}

	waitGroup.Wait()

	return preparedDotfiles
}
