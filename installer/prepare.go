package installer

import (
	"os"
	"sync"

	"go.evanpurkhiser.com/dots/config"
	"go.evanpurkhiser.com/dots/resolver"
)

// A PreparedDotfile represents a dotfile that has been "prepared" for
// installation by verifying it's contents against the existing dotfile, and
// checking various other flags that require knowledge of the existing dotfile.
type PreparedDotfile struct {
	*resolver.Dotfile

	// IsNew indicates that the dotfile does not currently exist.
	IsNew bool

	// ContentsDiffer is a boolean flag representing that the compiled source
	// differs from the currently installed dotfile.
	ContentsDiffer bool

	// SourcesAreIrregular indicates that a compiled dotfile (one with multiple
	// sources) does not have all regular file sources. This dotfile cannot be
	// compiled or installed.
	SourcesAreIrregular bool

	// Mode represents the mode bits of the os.FileMode (does not include
	// permissions). This will not be set if the sources are irregular.
	Mode os.FileMode

	// Permissions represents the change in permission between the compiled source
	// and the currently installed dotfile. Equal permissions can be verified
	// by calling Permissions.IsSame.
	Permissions Permissions

	// SourcePermissionsDiffer indicates that a compiled dotfile (one with
	// multiple sources) does not have consistent permissions across all
	// sources. In this case the lowest mode will be used.
	SourcePermissionsDiffer bool

	// RemovedNull is a warning flag indicating that the removed dotfile does
	// not exist in the install tree, though the dotfile is marked as removed.
	RemovedNull bool

	// OverwritesExisting is a warning flag that indicates that installing this
	// dotfile is overwriting a dotfile that was not part of the lockfile.
	OverwritesExisting bool

	// PrepareError keeps track of errors while preparing the dotfile. Should
	// this contain any errors, the PreparedDotfile is likely incomplete.
	PrepareError error

	sourceInfo []os.FileInfo
}

// Permissions represents a change in file permissions.
type Permissions struct {
	Old os.FileMode
	New os.FileMode
}

// IsSame returns a boolean value indicating if the modes are equal.
func (d Permissions) IsSame() bool {
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
			Dotfile: dotfile,
		}
		preparedDotfiles[index] = &prepared

		targetInfo, targetStatErr := os.Lstat(installPath)

		exists := !os.IsNotExist(targetStatErr)
		prepared.IsNew = exists

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

		prepared.sourceInfo = sourceInfo

		for i, source := range dotfile.Sources {
			path := config.SourcePath + separator + source.Path

			info, err := os.Lstat(path)
			if err != nil {
				prepared.PrepareError = err
				return
			}
			sourceInfo[i] = info
		}

		sourcePermissions, tookLowest := flattenPermissions(sourceInfo)

		prepared.Permissions = Permissions{
			Old: targetInfo.Mode() & os.ModePerm,
			New: sourcePermissions,
		}
		prepared.SourcePermissionsDiffer = tookLowest

		prepared.SourcesAreIrregular = !isAllRegular(sourceInfo)

		// NOTE: This currently drops non-mode bits (permissions are included
		// seperately), however should the dotfile set ModeAppend, ModeSticky,
		// etc, these modes will not be included here.
		if !prepared.SourcesAreIrregular {
			prepared.Mode = sourceInfo[0].Mode() & os.ModeType
		}

		// If the dotfile does not require compilation we can directly compare
		// the size of the (single) source file with the current target source.
		// Otherwise we will have to compile to compare.
		if !shouldCompile(dotfile, config) && targetInfo.Size() != sourceInfo[0].Size() {
			prepared.ContentsDiffer = true
			return
		}

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
