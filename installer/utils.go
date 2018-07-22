package installer

import (
	"bytes"
	"io"
	"os"
)

// flattenPermissions takes a list of objects implementing the os.FileInfo
// interface and flattens the permissions into a single FileMode. If any of the
// permissions differ, it will flatten the permission to into the lowest (least
// permissive) permission and set the boolean return value to true.
func flattenPermissions(infos []os.FileInfo) (m os.FileMode, tookLowest bool) {
	if len(infos) < 1 {
		return 0, false
	}

	lowestMode := infos[0].Mode() & os.ModePerm

	for _, stat := range infos[1:] {
		mode := stat.Mode() & os.ModePerm

		if mode != lowestMode {
			tookLowest = true
		}

		if mode < lowestMode {
			lowestMode = mode
		}
	}

	return lowestMode, tookLowest
}

// isAllRegular checks that a list of objects implementing the os.FileInfo
// interface all contain regular-type modes.
func isAllRegular(infos []os.FileInfo) bool {
	allRegular := true

	for _, info := range infos {
		if !info.Mode().IsRegular() {
			allRegular = false
			break
		}
	}

	return allRegular
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
