package installer

import (
	"bytes"
	"io"
	"os"
)

// flattenModes takes a list of objects implementing the os.FileInfo interface
// and flattens the mode into a single FileMode. If any of the modes differ, it
// will flatten the mode to into the lowest (least permissive) mode and set the
// boolean return value to true.
func flattenModes(infos []os.FileInfo) (m os.FileMode, tookLowest bool) {
	if len(infos) < 1 {
		return 0, false
	}

	lowestMode := infos[0].Mode()

	for _, stat := range infos[1:] {
		mode := stat.Mode()

		if mode != lowestMode {
			tookLowest = true
		}

		if mode < lowestMode {
			lowestMode = mode
		}
	}

	return lowestMode, tookLowest
}

const chunkSize = 4096

// compareReaders compares two io.Readers for differences.
func compareReaders(file1, file2 io.Reader) (bool, error) {
	b1 := make([]byte, chunkSize)
	b2 := make([]byte, chunkSize)

	for {
		n1, err1 := file1.Read(b1)
		n2, err2 := file2.Read(b2)

		if err1 != nil && err1 != io.EOF {
			return false, err1
		}
		if err2 != nil && err2 != io.EOF {
			return false, err2
		}

		if err1 == io.EOF || err2 == io.EOF {
			return err1 == err2, nil
		}

		if !bytes.Equal(b1[:n1], b2[:n2]) {
			return false, nil
		}
	}
}
