package resolver

import (
	"os"
	"path/filepath"
	"sort"
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

// Files gets the list of all configuration files.
func (c Configurations) Files() []string {
	files := []string{}

	for path := range c {
		files = append(files, path)
	}

	sort.Strings(files)

	return files
}

// Filter filters down a Configurations list to only configs with the specified prefix
func (c Configurations) Filter(prefix string) Configurations {
	for path := range c {
		if !strings.HasPrefix(path, prefix) {
			delete(c, path)
		}
	}

	return c
}

// resolveSources inserts or updates a Configurations list from a list of
// configuration sources relative to the source root. Sources not belonging to
// the specified group will be ignored.
func resolveSources(configs Configurations, sources []string, group string) {
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
			Sources: []*SourceFile{sourceFile},
		}
	}
}

// resolveOverrides scans a Configurations list for config files that have
// associated override files. The override files will be removed as individual
// configurations and their sources will be inserted into config file they are
// associated to.
func resolveOverrides(configs Configurations, overrideSuffix string) {
	for path, config := range configs {
		if strings.HasSuffix(path, "."+overrideSuffix) {
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
		delete(configs, overridePath)
	}
}

// resolveRemoved inserts entries into a Configurations list for files that
// previously were installed but are no longer present to be installed.
func resolveRemoved(configs Configurations, oldConfigs []string) {
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

// Config specifies the configuration object to be pased into the
// ResolveConfigurations function.
type Config struct {
	SourcePath     string
	OverrideSuffix string

	// Groups to resolve
	Groups []string

	// CurrentConfigs is a list of files that are currently installed. This
	// list will be used to determine which configurations have been removed.
	CurrentConfigs []string

	// FilterPrefix specifies the prefix that should be used to filter down the
	// list of installable configurations.
	FilterPrefix string
}

// ResolveConfigurations walks the source tree and builds a Configuration
// object, resolving group sources along the way.
func ResolveConfigurations(config Config) Configurations {
	sources := []string{}
	configs := Configurations{}

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		srcPath := strings.TrimPrefix(path, config.SourcePath+separator)
		sources = append(sources, srcPath)

		return nil
	}

	filepath.Walk(config.SourcePath, walker)

	for _, group := range config.Groups {
		resolveSources(configs, sources, group)
	}

	//resolveRemoved(configs, config.CurrentConfigs)
	resolveOverrides(configs, config.OverrideSuffix)

	return configs
}
