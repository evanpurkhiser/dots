package installer

import (
	"bytes"
	"os"
	"regexp"
)

var shebangRegex = regexp.MustCompile("^#!.*\n")

// trimShebang removes leading shebangs from a byte slice.
func trimShebang(d []byte) []byte {
	return shebangRegex.ReplaceAll(d, []byte{})
}

// trimWhitespace removes whitespace before and after a byte slice.
func trimWhitespace(d []byte) []byte {
	return bytes.Trim(d, "\n\t ")
}

var envGetter = os.Getenv

// expandEnvironment replaces environment variables.
func expandEnvironment(d []byte) []byte {
	return []byte(os.Expand(string(d), envGetter))
}
