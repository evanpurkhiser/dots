package resolver

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.evanpurkhiser.com/dots/config"
)

const separator = string(filepath.Separator)

// SourceFile represents a file that is used to compile a dotfile. The source
// file knows what group it is part of.
type SourceFile struct {
	Group    string
	Path     string
	Override bool
}

// Dotfile represents a file to be installed.
type Dotfile struct {
	Path    string
	Removed bool
	Sources []*SourceFile
}

// Dotfiles holds a list of Dotfiles.
type Dotfiles []*Dotfile

// Files gets the list of all configuration files.
func (d Dotfiles) Files() []string {
	files := []string{}

	for _, dotfile := range d {
		path := dotfile.Path
		files = append(files, path)
	}

	return files
}

// Filter filters down a Dotfiles list to only dotfiles with the specified
// prefixes.
func (d Dotfiles) Filter(prefixes []string) Dotfiles {
	if len(prefixes) == 0 {
		return d
	}

	dotfiles := Dotfiles{}

	for _, dotfile := range d {
		path := dotfile.Path
		filtered := true

		for _, prefix := range prefixes {
			if strings.HasPrefix(path, prefix) {
				filtered = false
				break
			}
		}

		if !filtered {
			dotfiles = append(dotfiles, dotfile)
		}

	}

	return dotfiles
}

// dotfiles is a package internal type used to construct the final list.
type dotfileMap map[string]*Dotfile

// asList constructs a Dotfiles object in order by the paths of the dotfiles.
func (d dotfileMap) asList() Dotfiles {
	paths := make([]string, 0, len(d))

	for path := range d {
		paths = append(paths, path)
	}

	sort.Strings(paths)

	dotfiles := Dotfiles{}

	for _, path := range paths {
		dotfiles = append(dotfiles, d[path])
	}

	return dotfiles
}

// resolveSources inserts or updates a dotfiles map from a list of
// dotfile sources relative to the source root. Sources not belonging to the
// specified group will be ignored.
func resolveSources(dotfiles dotfileMap, sources []string, group string) {
	for _, source := range sources {
		if !strings.HasPrefix(source, group) {
			continue
		}

		destPath := strings.TrimPrefix(source, group+separator)

		sourceFile := &SourceFile{
			Group: group,
			Path:  source,
		}

		// The dotfle file was added in a previous group mapping
		if file, ok := dotfiles[destPath]; ok {
			file.Sources = append(file.Sources, sourceFile)
			continue
		}

		dotfiles[destPath] = &Dotfile{
			Path:    destPath,
			Sources: []*SourceFile{sourceFile},
		}
	}
}

// resolveOverrides scans a dotfiles map for files that have associated
// override files. The override files will be removed as individual dotfiles
// and their sources will be inserted into the file they are associated to. Any
// override dotfiels which do not override an existing dotfile will have a new
// dotfile group created.
func resolveOverrides(dotfiles dotfileMap, overrideSuffix string) {
	flatten := map[string]string{}

	for path, dotfile := range dotfiles {
		if strings.HasSuffix(path, overrideSuffix) {
			// If this override file has nothing to override mark it to be
			// flattened (the suffix will be removed). We mark it instead of
			// flattening it into the dotfiles map here as items added to maps
			// may be itterated over if added while iterrating a map.
			overridesPath := strings.TrimSuffix(path, overrideSuffix)

			// Only if this override overides nothing
			if _, ok := dotfiles[overridesPath]; !ok {
				flatten[path] = overridesPath
			}

			continue
		}

		overridePath := path + overrideSuffix
		overrideDotfile, exists := dotfiles[overridePath]

		if !exists {
			continue
		}

		for _, source := range overrideDotfile.Sources {
			source.Override = true
		}

		dotfile.Sources = append(dotfile.Sources, overrideDotfile.Sources...)
		delete(dotfiles, overridePath)
	}

	for path, overridesPath := range flatten {
		dotfiles[overridesPath] = dotfiles[path]
		delete(dotfiles, path)

		file := dotfiles[overridesPath]

		// We can expect that only one source file should exist in the dotfiles
		// source set. Mark it as an override source
		file.Sources[0].Override = true
		file.Path = strings.TrimSuffix(file.Path, overrideSuffix)
	}
}

// resolveRemoved inserts entries into a dotfiles map for files that previously
// were installed but are no longer present to be installed.
func resolveRemoved(dotfiles dotfileMap, oldDotfiles []string) {
	for _, oldDotfile := range oldDotfiles {
		if _, ok := dotfiles[oldDotfile]; ok {
			continue
		}

		dotfiles[oldDotfile] = &Dotfile{
			Path:    oldDotfile,
			Removed: true,
		}
	}
}

// sourceLoader provides a list of files given a source path.
var sourceLoader = func(path string) []string {
	sources := []string{}

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		srcPath := strings.TrimPrefix(path, path+separator)
		sources = append(sources, srcPath)

		return nil
	}

	filepath.Walk(path, walker)

	return sources
}

// ResolveDotfiles walks the source tree and builds a Dotfiles object,
// resolving group sources along the way.
func ResolveDotfiles(conf config.SourceConfig, lockfile config.SourceLockfile) Dotfiles {
	dotfiles := dotfileMap{}

	sources := sourceLoader(conf.SourcePath)
	groups := lockfile.ResolveGroups(conf)

	for _, group := range groups {
		resolveSources(dotfiles, sources, group)
		resolveOverrides(dotfiles, "."+conf.OverrideSuffix)
	}

	resolveRemoved(dotfiles, lockfile.InstalledFiles)

	return dotfiles.asList()
}
