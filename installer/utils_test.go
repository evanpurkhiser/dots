package installer

import (
	"os"
	"testing"
	"time"
)

type modeStub os.FileMode

func (m modeStub) Name() string       { return "" }
func (m modeStub) Size() int64        { return 0 }
func (m modeStub) Mode() os.FileMode  { return os.FileMode(m) }
func (m modeStub) ModTime() time.Time { return time.Time{} }
func (m modeStub) IsDir() bool        { return false }
func (m modeStub) Sys() interface{}   { return nil }

func TestFlattenModes(t *testing.T) {
	testCases := []struct {
		caseName         string
		modes            []os.FileMode
		expectedMode     os.FileMode
		shouldTakeLowest bool
	}{
		{
			caseName: "All same permissions",
			modes: []os.FileMode{
				os.ModePerm & 0777,
				os.ModePerm & 0777,
				os.ModePerm & 0777,
			},
			expectedMode:     os.ModePerm & 0777,
			shouldTakeLowest: false,
		},
		{
			caseName: "Differing permissions",
			modes: []os.FileMode{
				os.ModePerm & 0755,
				os.ModePerm & 0644,
				os.ModePerm & 0777,
			},
			expectedMode:     os.ModePerm & 0644,
			shouldTakeLowest: true,
		},
		{
			caseName: "With extra file modes",
			modes: []os.FileMode{
				os.ModePerm&0777 | os.ModeDir,
				os.ModePerm&0644 | os.ModeCharDevice,
				os.ModePerm&0644 | os.ModeDir,
			},
			expectedMode:     os.ModePerm&0644 | os.ModeCharDevice,
			shouldTakeLowest: true,
		},
	}

	for _, testCase := range testCases {
		infos := make([]os.FileInfo, len(testCase.modes))

		for i, mode := range testCase.modes {
			infos[i] = modeStub(mode)
		}

		mode, tookLowest := flattenModes(infos)

		if mode != testCase.expectedMode {
			t.Errorf("Expected mode = %s; got mode = %s", mode, testCase.expectedMode)
		}

		if tookLowest != testCase.shouldTakeLowest {
			t.Errorf("Expected tookLowest = %t", testCase.shouldTakeLowest)
		}
	}
}
