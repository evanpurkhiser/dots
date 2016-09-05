package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const separator = string(filepath.Separator)

// SourceFile represents a file that is used to compile a configuration file.
// The source file knows what group it is part of.
type SourceFile struct {
	Group    string
	Path     string
	Override bool
}

// ConfigFile represents a configuration file to be installed.
type ConfigFile struct {
	Path    string
	Removed bool
	Sources []*SourceFile
}

// Configurations holds a mapping of configuration file destinations mapped to
// the ConfigFile struct representing that file.
type Configurations map[string]*ConfigFile

// ResolveSources inserts or updates a Configurations list from a list of
// configuration sources relative to the source root. Sources not belonging to
// the specified group will be ignored.
func ResolveSources(configs Configurations, sources []string, group string) {
	for _, source := range sources {
		if !strings.HasPrefix(source, group) {
			continue
		}

		destPath := strings.TrimPrefix(source, group+separator)

		sourceFile := &SourceFile{
			Group: group,
			Path:  source,
		}

		// The config file was added in a previous group mapping
		if file, ok := configs[destPath]; ok {
			file.Sources = append(file.Sources, sourceFile)
			continue
		}

		configs[destPath] = &ConfigFile{
			Path:    destPath,
			Sources: []SourceFile{sourceFile},
		}
	}
}

// ResolveOverrides scans a Configurations list for config files that have
// associated override files. The override files will be removed as individual
// configurations and their sources will be inserted into config file they are
// associated to.
func ResolveOverrides(configs Configurations, overrideSuffix string) {
	for path, config := range configs {
		if strings.HasSuffix(path, overrideSuffix) {
			continue
		}

		overridePath := path + overrideSuffix
		overrideConfig, exists := configs[overridePath]

		if !exists {
			continue
		}

		for _, source := range overrideConfig.Sources {
			source.Override = true
		}

		config.Sources = append(config.Sources, overrideConfig.Sources...)
		delete(config, overridePath)
	}
}

// ResolveRemoved inserts entries into a Configurations list for files that
// previously were installed but are no longer present to be installed.
func ResolveRemoved(configs Configurations, oldConfigs []string) {
	for _, oldConfig := range oldConfigs {
		if _, ok := configs[oldConfig]; ok {
			continue
		}

		configs[oldConfig] = &ConfigFile{
			Path:    oldConfig,
			Removed: true,
		}
	}
}

// ResolveScripts herp derp
func ResolveScripts(configs Configurations) {

}

func main() {
	// TODO: Verify config file exists
	configPath := "/home/evan/.local/etc"

	var sources []string

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		srcPath := strings.TrimPrefix(path, configPath+separator)
		sources = append(sources, srcPath)

		return nil
	}

	filepath.Walk(configPath, walker)

	configs := Configurations{}

	// TODO: Add group validation: ensure mutual exlcusion and files exist.
	//       Could just reduce the groups list and produce warnings for bad
	//       groups.
	groups := []string{
		"base",
		"machines/desktop",
	}

	for _, group := range groups {
		ResolveSources(configs, sources, group)
	}

	for _, config := range configs {
		fmt.Printf("%#v\n", config)
	}

	// TODO: Output writer that looks somthing like
	//
	// source:  /home/.local/etc
	// install: /home/.config
	//
	// [=> base ]    bash/bashrc
	// [=> composed] bash/environment
	//  -> composing from base and machines/desktop groups
	// [=> removed ] bash/bad-filemulti
	// [=> compiled] bash/complex
	//  -> ignoring configs in base and common/work due to override
	//  -> override file present in common/work-vm
	//  -> composing from machines/crunchydev (spliced at common/work-vm:22)
	//
	// [=> install script] base/bash.install
	//  -> triggered by base/bash/bashrc
	//  -> triggered by base/bash/environment

}
