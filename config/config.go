package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// DefaultConfigPath specifies the default path to look for source config file.
const defaultConfigPath = "${HOME}/.local/etc/config.yml"

// SourceConfigEnv is the name of the environment variable that will be read to
// override the location of the source config file.
const sourceConfigEnv = "DOTS_CONFIG"

// Profiles specifies the mapping of profile names to a list of group names
type Profiles map[string][]string

// Names returns a list of configured profiles
func (p Profiles) Names() []string {
	names := []string{}

	for name := range p {
		names = append(names, name)
	}

	return names
}

// SourceConfig specifies the structure of the source dotfile configuration.
type SourceConfig struct {
	// SourcePath specifies where the source configuration files live. If left
	// blank directory containing the source config will be used.
	SourcePath string `yaml:"source_path"`

	// InstallPath specifies where configuration files should be installed
	InstallPath string `yaml:"install_path"`

	// LockfilePath specifies where the dots lockfile should be installed. If
	// left blank it will be installed into ${install_path}/dots/lockfile.json
	LockfilePath string `yaml:"lockfile_path"`

	// OverrideSuffix specifies the file suffix to mark a configuration file as
	// overriding all files lower in the cascade.
	OverrideSuffix string `yaml:"override_suffix"`

	// InstallSuffix specifies the file suffix to mark a file as an
	// installation file.
	InstallSuffix string `yaml:"install_suffix"`

	// Groups specifies the configuration groups provided in the source
	// repository. These may be more than one directory level deep.
	Groups []string `yaml:"groups"`

	// BaseGroups specifies a list of groups that should always be installed.
	BaseGroups []string `yaml:"base_groups"`

	// Profiles is a mapping of profile names to a list of groups to install.
	// Base groups do not need to be specified.
	Profiles Profiles `yaml:"profiles"`

	// ExpandEnvironment specifies a list of install file paths that when
	// installed should have bash style parameter expansion done for
	// environment variables. These generally look like: ${HOME}. This
	// configuration is useful for config files that do not support environment
	// variable expansion.
	ExpandEnvironment []string `yaml:"expand_environment"`
}

// SourceLockfile specifies the structure of the lockfile that is installed
// along side configuration files.
type SourceLockfile struct {
	// Profile specifies the currently configured profile
	Profile string `json:"profile"`

	// Groups is a list of groups to be installed if no profile is specified.
	// Does nothing if a profile has been specified. The base groups from the
	// configuration will always be installed.
	Groups []string `json:"groups"`

	// IntalledFiles is the current list of insatlled configuration files
	InstalledFiles []string `json:"installed_files"`
}

// ResolveGroups returns the list of groups the lockfile is currently locked to.
func (l *SourceLockfile) ResolveGroups(config SourceConfig) []string {
	profileGroups, ok := config.Profiles[l.Profile]
	if ok {
		return append(config.BaseGroups, profileGroups...)
	}

	// Use lockfile groups override list
	return append(config.BaseGroups, l.Groups...)
}

// SourceConfigPath determines the location of the source configuration file.
func SourceConfigPath() string {
	if path := os.Getenv(sourceConfigEnv); path != "" {
		return path
	}

	return os.ExpandEnv(defaultConfigPath)
}

// LoadSourceConfig reads and unmarshals the yaml source configuration file
// into the SourceConfig.
func LoadSourceConfig(configPath string) (*SourceConfig, error) {
	config := &SourceConfig{}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	// Determine the source path if not configured
	if config.SourcePath == "" {
		config.SourcePath = filepath.Dir(configPath)
	}

	// Resolve environment variables
	config.SourcePath = os.ExpandEnv(config.SourcePath)
	config.InstallPath = os.ExpandEnv(config.InstallPath)
	config.LockfilePath = os.ExpandEnv(config.LockfilePath)

	// Determine the lockfile path if not configured
	if config.LockfilePath == "" {
		config.LockfilePath = path.Join(config.InstallPath, "dots", "dotlock.json")
	}

	return config, nil
}

// LoadLockfile reads and unmarshals the json lockfile into a SourceLockfile.
func LoadLockfile(config *SourceConfig) (*SourceLockfile, error) {
	lockfile := &SourceLockfile{}

	file, err := os.Open(config.LockfilePath)
	if err != nil {
		return lockfile, nil
	}

	defer file.Close()

	if err := json.NewDecoder(file).Decode(lockfile); err != nil {
		return nil, err
	}

	return lockfile, nil
}

// WriteLockfile updates or creates the lockfile on the system.
func WriteLockfile(lockfile *SourceLockfile, config *SourceConfig) error {
	if err := os.MkdirAll(path.Dir(config.LockfilePath), 0777); err != nil {
		return err
	}

	file, err := os.OpenFile(config.LockfilePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	defer file.Close()

	if err := json.NewEncoder(file).Encode(lockfile); err != nil {
		return err
	}

	return file.Sync()
}
