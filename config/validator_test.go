package config

import (
	"os"
	"path/filepath"
	"testing"
)

var tmpPath = os.TempDir()

// TestMissingSourceDir tests that if the source directory does not exist we
// receive exactly one error.
func TestMissingSourceDir(t *testing.T) {
	config := &SourceConfig{
		SourcePath: filepath.Join(tmpPath, "path-does-not-exist"),
	}

	errs := SanitizeSourceConfig(config)

	if len(errs) != 1 {
		t.Fatalf("Expected len(errs) = 1; got %d", len(errs))
	}

	if errs[0].Error() != "Specified source_path does not exist" {
		t.Errorf("Expected source_path does not exist error message")
	}
}

// TestRemoveDuplicateGroups tests that duplicate groups are removed and a
// warning is produced.
func TestRemoveDuplicateGroups(t *testing.T) {
	config := &SourceConfig{
		SourcePath: tmpPath,
		Groups: []string{
			"group1",
			"group1",
		},
	}

	groupPath := filepath.Join(tmpPath, "group1")

	os.Mkdir(groupPath, 0755)
	defer os.Remove(groupPath)

	errs := SanitizeSourceConfig(config)

	if len(errs) != 1 {
		t.Fatalf("Expected len(errs) = 1; got %d", len(errs))
	}

	if errs[0].Error() != "Group \"group1\" already specified" {
		t.Errorf("Expected duplicate error; got = %q", errs[0])
	}

	if len(config.Groups) != 1 {
		t.Errorf("Expected duplicate to be removed from groups list")
	}
}

// TestRemoveMissingGroups tests that groups that do not exist as directories
// on the filesystem are removed from the groups list along with base groups
// list and all profile groups lists.
func TestRemoveMissingGroups(t *testing.T) {
	config := &SourceConfig{
		SourcePath: tmpPath,
		Groups:     []string{"group1", "group2"},
		BaseGroups: []string{"group1"},

		Profiles: Profiles{
			"test-profile": []string{"group2"},
		},
	}

	errs := SanitizeSourceConfig(config)

	if len(errs) != 2 {
		t.Fatalf("Expected len(errs) = 2; got %d", len(errs))
	}

	if errs[0].Error() != "Group \"group1\" does not exist in sources" {
		t.Errorf("Expected missing error; got = %q", errs[0])
	}

	if len(config.Groups) != 0 {
		t.Errorf("Expected missing group to be removed from groups list")
	}

	if len(config.BaseGroups) != 0 {
		t.Errorf("Expected missing group to be removed from base groups list")
	}

	if len(config.Profiles["test-profile"]) != 0 {
		t.Errorf("Expected missing group to be removed from base test-profile groups")
	}
}

// TestBaseGroupsMustBeValid tests that base groups which do not exist in the
// main groups list are removed.
func TestBaseGroupsMustBeValid(t *testing.T) {
	config := &SourceConfig{
		SourcePath: tmpPath,
		Groups:     []string{"group1"},
		BaseGroups: []string{"group2"},
	}

	groupPath := filepath.Join(tmpPath, "group1")

	os.Mkdir(groupPath, 0755)
	defer os.Remove(groupPath)

	errs := SanitizeSourceConfig(config)

	if len(errs) != 1 {
		t.Fatalf("Expected len(errs) = 1; got %d", len(errs))
	}

	if errs[0].Error() != "Base group \"group2\" is not a valid group" {
		t.Errorf("Expected missing error; got = %q", errs[0])
	}

	if len(config.BaseGroups) != 0 {
		t.Errorf("Expected bad group to be removed from base groups list")
	}
}

// TestBaseGroupDuplicatesRemoved tests that duplicates groups in the base
// groups list are removed.
func TestBaseGroupDuplicatesRemoved(t *testing.T) {
	config := &SourceConfig{
		SourcePath: tmpPath,
		Groups:     []string{"group1"},
		BaseGroups: []string{"group1", "group1"},
	}

	groupPath := filepath.Join(tmpPath, "group1")

	os.Mkdir(groupPath, 0755)
	defer os.Remove(groupPath)

	errs := SanitizeSourceConfig(config)

	if len(errs) != 1 {
		t.Fatalf("Expected len(errs) = 1; got %d", len(errs))
	}

	if errs[0].Error() != "Base group \"group1\" already specified" {
		t.Errorf("Expected missing error; got = %q", errs[0])
	}

	if len(config.BaseGroups) != 1 {
		t.Errorf("Expected duplicate group to be removed from base groups list")
	}
}

// TestInvalidProfileGroup tests that profiles with groups that do not exist
// have the groups removed.
func TestInvalidProfileGroup(t *testing.T) {
	config := &SourceConfig{
		SourcePath: tmpPath,
		Groups:     []string{"group1"},
		Profiles: Profiles{
			"test-profile":  []string{"group1"},
			"test-profile2": []string{"group1", "group2"},
		},
	}

	groupPath := filepath.Join(tmpPath, "group1")

	os.Mkdir(groupPath, 0755)
	defer os.Remove(groupPath)

	errs := SanitizeSourceConfig(config)

	if len(errs) != 1 {
		t.Fatalf("Expected len(errs) = 1; got %d", len(errs))
	}

	if errs[0].Error() != "Profile \"test-profile2\": Group \"group2\" is not a valid group" {
		t.Errorf("Expected missing error; got = %q", errs[0])
	}

	if len(config.Profiles["test-profile2"]) != 1 {
		t.Errorf("Expected invalid group to be removed from test-profile2")
	}
}

// TestProfileDuplicateGroups tests that profiles with duplicated groups have
// the groups removed.
func TestProfileDuplicateGroups(t *testing.T) {
	config := &SourceConfig{
		SourcePath: tmpPath,
		Groups:     []string{"group1"},
		Profiles: Profiles{
			"test-profile":  []string{"group1"},
			"test-profile2": []string{"group1", "group1"},
		},
	}

	groupPath := filepath.Join(tmpPath, "group1")

	os.Mkdir(groupPath, 0755)
	defer os.Remove(groupPath)

	errs := SanitizeSourceConfig(config)

	if len(errs) != 1 {
		t.Fatalf("Expected len(errs) = 1; got %d", len(errs))
	}

	if errs[0].Error() != "Profile \"test-profile2\": Group \"group1\" already specified" {
		t.Errorf("Expected missing error; got = %q", errs[0])
	}

	if len(config.Profiles["test-profile2"]) != 1 {
		t.Errorf("Expected duplicate group to be removed from test-profile2")
	}
}

// TestProfileRemoveBaseGroups tests that profiles with base groups have
// the groups removed.
func TestProfileRemoveBaseGroups(t *testing.T) {
	config := &SourceConfig{
		SourcePath: tmpPath,
		Groups:     []string{"group1"},
		BaseGroups: []string{"group1"},
		Profiles: Profiles{
			"test-profile": []string{"group1"},
		},
	}

	groupPath := filepath.Join(tmpPath, "group1")

	os.Mkdir(groupPath, 0755)
	defer os.Remove(groupPath)

	errs := SanitizeSourceConfig(config)

	if len(errs) != 1 {
		t.Fatalf("Expected len(errs) = 1; got %d", len(errs))
	}

	if errs[0].Error() != "Profile \"test-profile\": Group \"group1\" is already specified in the base groups" {
		t.Errorf("Expected missing error; got = %q", errs[0])
	}

	if len(config.Profiles["test-profile"]) != 0 {
		t.Errorf("Expected base group to be removed from test-profile")
	}
}

// TestValidateLockfile tests that we validate a valid lockfile without errors.
func TestValidateLockfile(t *testing.T) {
	config := &SourceConfig{
		Profiles: Profiles{"test-profile": []string{}},
	}

	lock := &SourceLockfile{
		Profile: "test-profile",
		Groups:  []string{},
	}

	// Valid profile-configured lockfile
	err := ValidateLockfile(lock, config)

	if err != nil {
		t.Errorf("Expected valid lockfile; got err = %q", err)
	}

	config = &SourceConfig{
		Groups:     []string{"group1", "group2"},
		BaseGroups: []string{"group1"},
	}

	lock = &SourceLockfile{
		Profile: "",
		Groups:  []string{"group2"},
	}

	// Valid group-configured lockfile
	err = ValidateLockfile(lock, config)

	if err != nil {
		t.Errorf("Expected valid lockfile; got err = %q", err)
	}
}

// TestInvalidLockfileProfile tests that a invalid profile in a lockfile
// results in a validation error.
func TestInvalidLockfileProfile(t *testing.T) {
	config := &SourceConfig{
		Profiles: Profiles{"test-profile": []string{}},
	}

	lock := &SourceLockfile{
		Profile: "test-profile2",
		Groups:  []string{},
	}

	err := ValidateLockfile(lock, config)

	if err == nil || err.Error() != "Profile \"test-profile2\" is not a configured profile" {
		t.Errorf("Expected invalid lockfile due to invalid profile; got err = %q", err)
	}
}

// TestLockfileHasGroupsAndProfile tests that a lockfile containing a profile
// and a list of groups results in a validation error.
func TestLockfileHasGroupsAndProfile(t *testing.T) {
	config := &SourceConfig{
		Profiles: Profiles{"test-profile": []string{}},
	}

	lock := &SourceLockfile{
		Profile: "test-profile",
		Groups:  []string{"group1"},
	}

	err := ValidateLockfile(lock, config)

	if err == nil || err.Error() != "Groups should not be specified if a profile is configured" {
		t.Errorf("Expected invalid lockfile due to groups and profile; got err = %q", err)
	}
}

// TestLockfileInvalidGroups tests that a lockfile containing a groups list
// with a group not specified in the source groups list results in a validation
// error.
func TestLockfileInvalidGroups(t *testing.T) {
	config := &SourceConfig{
		Groups: []string{"group1"},
	}

	lock := &SourceLockfile{
		Groups: []string{"group1", "group2"},
	}

	err := ValidateLockfile(lock, config)

	if err == nil || err.Error() != "Lockfile contains invalid groups" {
		t.Errorf("Expected invalid lockfile due to invalid groups; got err = %q", err)
	}
}

// TestLockfileContainsBaseGroups tests that a lockfile containing a groups
// list including configured base groups results in a validation error.
func TestLockfileContainsBaseGroups(t *testing.T) {
	config := &SourceConfig{
		Groups:     []string{"group1", "group2"},
		BaseGroups: []string{"group1"},
	}

	lock := &SourceLockfile{
		Groups: []string{"group1", "group2"},
	}

	err := ValidateLockfile(lock, config)

	if err == nil || err.Error() != "Lockfile groups includes a base group" {
		t.Errorf("Expected invalid lockfile due to included base groups; got err = %q", err)
	}
}
