package resolver

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"go.evanpurkhiser.com/dots/config"
)

func TestResolveDotfiles(t *testing.T) {
	tests := []struct {
		CaseName       string
		SourceFiles    []string
		ExistingFiles  []string
		Groups         []string
		OverrideSuffix string
		Expected       Dotfiles
	}{
		{
			CaseName:      "Noop, no source files to install, produce no dotfiles.",
			SourceFiles:   []string{},
			ExistingFiles: []string{},
			Groups:        []string{},
			Expected:      Dotfiles{},
		},
		{
			CaseName:      "Noop, files without any groups.",
			SourceFiles:   []string{"base/one", "machine/two"},
			ExistingFiles: []string{},
			Groups:        []string{},
			Expected:      Dotfiles{},
		},
		// Combinative test. Includes a single file per group, files composed
		// from groups, a override file in groups.
		{
			CaseName: "Combined common test",
			SourceFiles: []string{
				"base/bash/conf",
				"base/environment",
				"base/multi-composed",
				"base/generic-config",
				"machines/desktop/generic-config.ovrd",
				"machines/desktop/vimrc",
				"machines/desktop/environment",
				"machines/server/multi-composed",
				"machines/server/blank.ovrd",
				"machines/desktop/multi-composed",
			},
			ExistingFiles:  []string{},
			Groups:         []string{"base", "machines/desktop", "machines/server"},
			OverrideSuffix: "ovrd",
			Expected: Dotfiles{
				{
					Path: "bash/conf",
					Sources: []*SourceFile{
						{
							Group: "base",
							Path:  "base/bash/conf",
						},
					},
				},
				{
					Path: "blank.ovrd",
					Sources: []*SourceFile{
						{
							Group: "machines/server",
							Path:  "machines/server/blank.ovrd",
						},
					},
				},
				{
					Path: "environment",
					Sources: []*SourceFile{
						{
							Group: "base",
							Path:  "base/environment",
						},
						{
							Group: "machines/desktop",
							Path:  "machines/desktop/environment",
						},
					},
				},
				{
					Path: "generic-config",
					Sources: []*SourceFile{
						{
							Group: "base",
							Path:  "base/generic-config",
						},
						{
							Group:    "machines/desktop",
							Path:     "machines/desktop/generic-config.ovrd",
							Override: true,
						},
					},
				},
				{
					Path: "multi-composed",
					Sources: []*SourceFile{
						{
							Group: "base",
							Path:  "base/multi-composed",
						},
						{
							Group: "machines/desktop",
							Path:  "machines/desktop/multi-composed",
						},
						{
							Group: "machines/server",
							Path:  "machines/server/multi-composed",
						},
					},
				},
				{
					Path: "vimrc",
					Sources: []*SourceFile{
						{
							Group: "machines/desktop",
							Path:  "machines/desktop/vimrc",
						},
					},
				},
			},
		},
		{
			CaseName:      "Dotfile removed",
			SourceFiles:   []string{"base/vimrc"},
			ExistingFiles: []string{"bashrc", "vimrc"},
			Groups:        []string{"base"},
			Expected: Dotfiles{
				{
					Path:    "bashrc",
					Removed: true,
				},
				{
					Path: "vimrc",
					Sources: []*SourceFile{
						{
							Group: "base",
							Path:  "base/vimrc",
						},
					},
				},
			},
		},
	}

	ogSourceLoader := sourceLoader
	defer func() { sourceLoader = ogSourceLoader }()

	for _, test := range tests {
		if test.OverrideSuffix == "" {
			test.OverrideSuffix = "override"
		}

		conf := config.SourceConfig{
			BaseGroups:     []string{},
			OverrideSuffix: test.OverrideSuffix,
		}

		lockfile := config.SourceLockfile{
			InstalledFiles: test.ExistingFiles,
			Groups:         test.Groups,
		}

		sourceLoader = func(path string) []string {
			return test.SourceFiles
		}

		dotfiles := ResolveDotfiles(conf, lockfile)

		if reflect.DeepEqual(dotfiles, test.Expected) {
			continue
		}

		dotfileList := []string{}

		for _, dotfile := range dotfiles {
			sourceFiles := []string{}
			for _, source := range dotfile.Sources {
				sourceFiles = append(sourceFiles, source.Path)
			}

			str := fmt.Sprintf("%s: [%s]", dotfile.Path, strings.Join(sourceFiles, ", "))
			dotfileList = append(dotfileList, str)
		}

		t.Errorf(
			"Mismatching dotfiles set.\nTest Case: %q\nGot files:\n%s",
			test.CaseName,
			strings.Join(dotfileList, "\n"),
		)
	}
}

func TestDotfilesFiles(t *testing.T) {
	dotfiles := Dotfiles{
		{
			Path: "bash/conf",
		},
		{
			Path: "environment",
		},
		{
			Path: "something/else",
		},
	}

	dotfileFiles := dotfiles.Files()

	expected := []string{
		"bash/conf",
		"environment",
		"something/else",
	}

	if !reflect.DeepEqual(dotfileFiles, expected) {
		t.Errorf("Expected = %v; got = %#v", expected, dotfileFiles)
	}
}

func TestDotfilesFilter(t *testing.T) {
	dotfiles := Dotfiles{
		{
			Path: "bash/conf",
		},
		{
			Path: "environment",
		},
		{
			Path: "something/else",
		},
		{
			Path: "something/else2",
		},
	}

	tests := []struct {
		prefixes []string
		expected Dotfiles
	}{
		{
			[]string{},
			dotfiles,
		},
		{
			[]string{"bash"},
			Dotfiles{dotfiles[0]},
		},
		{
			[]string{"something"},
			Dotfiles{dotfiles[2], dotfiles[3]},
		},
		{
			[]string{"invalid"},
			Dotfiles{},
		},
		{
			[]string{"bash", "something"},
			Dotfiles{dotfiles[0], dotfiles[2], dotfiles[3]},
		},
	}

	for _, test := range tests {
		filtered := dotfiles.Filter(test.prefixes)

		if !reflect.DeepEqual(filtered, test.expected) {
			t.Errorf("Expected = %v; got = %#v", test.expected, filtered)
		}
	}
}
