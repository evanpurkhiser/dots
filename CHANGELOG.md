# Changelog

## [2.0.0] - TBA

Version 2 of the `dots` tool is a complete rewrite from the original single file
Python script, into a Go binary. This rewrite pairs down the features,
introduces verbose output, adds dry-runs, more robust error handling, and
improved configuration.

### Changed

- Rewritten in Go.

### Added

- `dots` has learned how to output verbose details about what the tool is doing.

- `dots` has learned how to perform a dry-run.

- `dots` has learned how to reinstall dotifles.

- `.install` files may now be named after directories, in which case they will
  be run any time a file within that directory is installed.

- Removed files will now also be correctly removed using a "lockfile" to keep
  track of what files were last expected to be installed.

- Configuration is now done in a `config.yml` file that lives in the root of
  your dots source directory.

- Environment variable expansion can now be performed as an installation time
  modification.

### Removed

- Support for explicit "default and "named" append points has been removed. The
  `!!@@` syntax has been removed. Dotfiles will now always be cascadded
  together linearly.

## [1.0.0] - 2014-01-03

- Initial release as a single file Python script.
