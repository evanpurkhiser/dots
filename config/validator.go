package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// SanitizeSourceConfig validates and sanitizes a SourceConfig object, ensuring
// the following conditions are met:
//
//  1. The source path exists.
//
//  2. Groups do not have any duplicates.
//  3. All groups exist as directories in the configured source path.
//  4. Base groups exist in the configured groups list.
//  5. Base groups do not have duplicates.
//  6. All profiles have groups that exist in the configured groups.
//  7. All profiles do not specify duplicate groups.
//  8. All profiles do not specify groups that are already configured as base groups.
//
// Any groups that do not meet these conditions will be removed from the group
// list being sanitized.
func SanitizeSourceConfig(config *SourceConfig) []error {
	errs := []error{}

	// 1. Ensure the source path exists
	if _, err := os.Stat(config.SourcePath); os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("Specified source_path does not exist"))

		// Most other validations will fail since the groups cannot be found.
		// Bail out now and deal with missing all sources later.
		return errs
	}

	config.SourcePath = filepath.Clean(config.SourcePath)
	config.InstallPath = filepath.Clean(config.InstallPath)

	// 2. Remove duplicate groups
	dedupedGroups, dupeGroups := removeDupes(config.Groups)
	config.Groups = dedupedGroups

	for _, group := range dupeGroups {
		errs = append(errs, fmt.Errorf("Group %q already specified", group))
	}

	// Do not sanitize the groups list for missing groups until all errors are
	// collected. This prevents many duplicate errors after removing a group
	// used in profiles / base groups.
	missingGroups := []string{}

	// 3. Locate missing groups
	for _, group := range config.Groups {
		path := filepath.Join(config.SourcePath, group)

		s, err := os.Stat(path)
		if !os.IsNotExist(err) && s.IsDir() {
			continue
		}

		errs = append(errs, fmt.Errorf("Group %q does not exist in sources", group))

		missingGroups = append(missingGroups, group)
	}

	// 4. Base groups must exist in the configured groups
	badBaseGroups := listDifference(config.BaseGroups, config.Groups)

	for _, group := range badBaseGroups {
		errs = append(errs, fmt.Errorf("Base group %q is not a valid group", group))
	}

	config.BaseGroups = listDifference(config.BaseGroups, badBaseGroups)

	// 5. Remove duplicate base groups
	dedupedBaseGroups, dupeBaseGroups := removeDupes(config.BaseGroups)
	config.BaseGroups = dedupedBaseGroups

	for _, group := range dupeBaseGroups {
		errs = append(errs, fmt.Errorf("Base group %q already specified", group))
	}

	for profile, groups := range config.Profiles {
		// 6. Profile groups must exist in configured groups
		badGroups := listDifference(groups, config.Groups)

		for _, group := range badGroups {
			err := fmt.Errorf("Profile %q: Group %q is not a valid group", profile, group)
			errs = append(errs, err)
		}

		config.Profiles[profile] = listDifference(groups, badGroups)
		groups = config.Profiles[profile]

		// 7. Profile does not specify duplicate groups.
		dedupedGroups, dupeGroups = removeDupes(groups)

		config.Profiles[profile] = dedupedGroups
		groups = config.Profiles[profile]

		for _, group := range dupeGroups {
			err := fmt.Errorf("Profile %q: Group %q already specified", profile, group)
			errs = append(errs, err)
		}

		// 8. Profile does not include a base group
		includedBaseGroups := listIntersect(groups, config.BaseGroups)

		for _, group := range includedBaseGroups {
			err := fmt.Errorf("Profile %q: Group %q is already specified in the base groups", profile, group)
			errs = append(errs, err)
		}

		config.Profiles[profile] = listDifference(groups, config.BaseGroups)
	}

	// Sanitize missing groups from group list, base groups, and all profiles
	config.Groups = listDifference(config.Groups, missingGroups)
	config.BaseGroups = listDifference(config.BaseGroups, missingGroups)

	for profile, groups := range config.Profiles {
		config.Profiles[profile] = listDifference(groups, missingGroups)
	}

	return errs
}

// ValidateLockfile validates the lockfile with the following conditions:
//
// If a profile is specified:
//
//  1. The profile should exist in the list of profiles
//  2. Groups should not also be specified.
//
// If groups are not blank:
//
//  3. Check that groups exists in the source config groups list.
//  4. Check that groups do not include any base groups.
//
// The lockfile should not be manually modified, so this is a sanity check,
// thus a lockfile failing validation should trigger a hard failure, asking the
// user to reconfigure their profile or groups.
//
// Unlike the SanitizeSourceConfig function, no modifications will be made to
// the lockfile struct.
func ValidateLockfile(lockfile *SourceLockfile, config *SourceConfig) error {
	// 1. Check that specified profile is valid
	if _, ok := config.Profiles[lockfile.Profile]; lockfile.Profile != "" && !ok {
		return fmt.Errorf("Profile %q is not a configured profile", lockfile.Profile)
	}

	// 2. Check that no groups are specified with a profile
	if lockfile.Profile != "" && len(lockfile.Groups) > 0 {
		return fmt.Errorf("Groups should not be specified if a profile is configured")
	}

	hasGroups := len(lockfile.Groups) > 0 && lockfile.Profile == ""

	// 3. Check that all groups are valid groups
	invalidGroups := listDifference(lockfile.Groups, config.Groups)

	if hasGroups && len(invalidGroups) > 0 {
		return fmt.Errorf("Lockfile contains invalid groups")
	}

	// 4. Check for base groups included in the groups
	baseGroups := listIntersect(lockfile.Groups, config.BaseGroups)

	if hasGroups && len(baseGroups) > 0 {
		return fmt.Errorf("Lockfile groups includes a base group")
	}

	return nil
}
