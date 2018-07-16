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
		InstallSuffix  string
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
				"base/bash/conf2",
				"base/bash.inst",
				"base/environment",
				"base/multi-composed",
				"base/generic-config",
				"base/generic-config.inst",
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
			InstallSuffix:  "inst",
			Expected: Dotfiles{
				{
					Path: "bash/conf",
					Sources: []*SourceFile{
						{
							Group: "base",
							Path:  "base/bash/conf",
						},
					},
					InstallFiles: []string{"base/bash.inst"},
				},
				{
					Path: "bash/conf2",
					Sources: []*SourceFile{
						{
							Group: "base",
							Path:  "base/bash/conf2",
						},
					},
					InstallFiles: []string{"base/bash.inst"},
				},
				{
					Path: "blank",
					Sources: []*SourceFile{
						{
							Group:    "machines/server",
							Path:     "machines/server/blank.ovrd",
							Override: true,
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
					InstallFiles: []string{"base/generic-config.inst"},
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
		// Override resolve order test. Override's should resolve as the groups
		// are resolved.
		{
			CaseName: "Override resolve order",
			SourceFiles: []string{
				"base/bash/file1",
				"machines/desktop/bash/file1.override",
				"machines/server/bash/file1",
				// Ensure resolving an override when there was no file to
				// override does not override _later_ groups.
				"base/bash/file2.override",
				"machines/desktop/bash/file2",
				// Ensure override files without any files to override have
				// their override stripped.
				"base/bash/file3.override",

				"base/bash/file4.override",
				"machines/desktop/bash/file4.override",
			},
			ExistingFiles: []string{},
			Groups:        []string{"base", "machines/desktop", "machines/server"},
			Expected: Dotfiles{
				{
					Path: "bash/file1",
					Sources: []*SourceFile{
						{
							Group: "base",
							Path:  "base/bash/file1",
						},
						{
							Group:    "machines/desktop",
							Path:     "machines/desktop/bash/file1.override",
							Override: true,
						},
						{
							Group: "machines/server",
							Path:  "machines/server/bash/file1",
						},
					},
				},
				{
					Path: "bash/file2",
					Sources: []*SourceFile{
						{
							Group:    "base",
							Path:     "base/bash/file2.override",
							Override: true,
						},
						{
							Group: "machines/desktop",
							Path:  "machines/desktop/bash/file2",
						},
					},
				},
				{
					Path: "bash/file3",
					Sources: []*SourceFile{
						{
							Group:    "base",
							Path:     "base/bash/file3.override",
							Override: true,
						},
					},
				},
				{
					Path: "bash/file4",
					Sources: []*SourceFile{
						{
							Group:    "base",
							Path:     "base/bash/file4.override",
							Override: true,
						},
						{
							Group:    "machines/desktop",
							Path:     "machines/desktop/bash/file4.override",
							Override: true,
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

		if test.InstallSuffix == "" {
			test.InstallSuffix = "install"
		}

		conf := config.SourceConfig{
			BaseGroups:     []string{},
			OverrideSuffix: test.OverrideSuffix,
			InstallSuffix:  test.InstallSuffix,
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
				if source.Override {
					sourceFiles = append(sourceFiles, source.Path+"[O]")
				} else {
					sourceFiles = append(sourceFiles, source.Path)
				}
			}

			str := fmt.Sprintf(
				"%s: [%s] I[%s]",
				dotfile.Path,
				strings.Join(sourceFiles, ", "),
				strings.Join(dotfile.InstallFiles, ", "),
			)

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
