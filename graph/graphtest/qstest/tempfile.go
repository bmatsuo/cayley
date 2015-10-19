package qstest

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/bmatsuo/cayley/graph/graphtest/testrunner"
)

type tmpFileKey struct{}
type tmpDirKey struct{}

// MkTempDir creates a temprorary directory using ioutil.TempDir and stores the
// path in vars.  Only one temporary directory can be stored in a Vars.
func MkTempDir(vars testrunner.Vars, tempdir, prefix string) (string, error) {
	tmpDir, err := ioutil.TempDir("", "cayley_test")
	if err != nil {
		return "", fmt.Errorf("temporary directory: %v", err)
	}
	vars[tmpDirKey{}] = tmpDir
	return tmpDir, nil
}

// RmTempDir removes the temporary directory stored in vars.
func RmTempDir(vars testrunner.Vars) {
	path, ok := vars[tmpDirKey{}].(string)
	if !ok {
		return
	}
	os.RemoveAll(path)
}

// MkTempFile creates a temprorary file using ioutil.TempFile and stores the
// path in vars.  Only one temporary file can be stored in a Vars.
func MkTempFile(vars testrunner.Vars, tempdir, prefix string) (*os.File, error) {
	f, err := ioutil.TempFile("", "cayley_test")
	if err != nil {
		return nil, fmt.Errorf("temporary directory: %v", err)
	}
	vars[tmpFileKey{}] = f.Name()
	return f, nil
}

// RmTempFile removes the temporary file stored in vars.
func RmTempFile(vars testrunner.Vars) {
	path, ok := vars[tmpFileKey{}].(string)
	if !ok {
		return
	}
	os.RemoveAll(path)
}
