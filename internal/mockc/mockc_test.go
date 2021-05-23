package mockc

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testRoot = "testdata"

type testCase struct {
	name    string
	path    string
	goFiles map[string][]byte

	input struct {
		Patterns []string
	}
	output struct {
		Output string
		Err    string
	}
	expectedFiles map[string][]byte
}

func TestMockc(t *testing.T) {
	a := assert.New(t)

	dirs, err := ioutil.ReadDir(testRoot)
	a.NoError(err)

	testCases := make([]testCase, 0, len(dirs))
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		tc := testCase{
			name:          dir.Name(),
			path:          filepath.Join(testRoot, dir.Name()),
			goFiles:       map[string][]byte{},
			expectedFiles: map[string][]byte{},
		}

		goFiles, err := filepath.Glob(filepath.Join(tc.path, "*.go"))
		a.NoError(err)
		for _, f := range goFiles {
			b, err := ioutil.ReadFile(f)
			a.NoError(err)
			tc.goFiles[filepath.Base(f)] = b
		}

		input, err := ioutil.ReadFile(filepath.Join(tc.path, "testdata", "input.json"))
		a.NoError(err)

		err = json.Unmarshal(input, &tc.input)
		a.NoError(err)

		output, err := ioutil.ReadFile(filepath.Join(tc.path, "testdata", "output.json"))
		a.NoError(err)

		err = json.Unmarshal(output, &tc.output)
		a.NoError(err)

		expectedFiles, err := filepath.Glob(filepath.Join(tc.path, "testdata", "*.go.gen"))
		a.NoError(err)
		for _, f := range expectedFiles {
			b, err := ioutil.ReadFile(f)
			a.NoError(err)
			tc.expectedFiles[strings.TrimSuffix(filepath.Base(f), ".gen")] = b
		}

		testCases = append(testCases, tc)
	}

	for _, tc := range testCases {
		testMockc(t, tc)
	}
}

func testMockc(t *testing.T, tc testCase) {
	t.Run(tc.name, func(t *testing.T) {
		a := assert.New(t)

		buf := bytes.NewBuffer(nil)
		log.SetFlags(0)
		log.SetOutput(buf)

		err := Generate(context.Background(), tc.path, tc.input.Patterns)
		if tc.output.Err == "" {
			a.NoError(err)
		} else {
			a.EqualError(err, tc.output.Err)
		}
		a.Regexp(regexp.MustCompile(tc.output.Output), buf.String())

		t.Cleanup(func() {
			for name := range tc.expectedFiles {
				err = os.Remove(filepath.Join(tc.path, name))
				a.NoError(err)
			}
		})

		for name, expected := range tc.expectedFiles {
			path := filepath.Join(tc.path, name)

			actual, err := ioutil.ReadFile(path)
			a.NoError(err)
			a.Equal(string(expected), string(actual))
		}
	})
}
