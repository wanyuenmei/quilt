package main

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"sort"

	"github.com/quilt/quilt/scripts/specs-tester/tests"
)

func main() {
	testsByName := map[string]func() error{
		"TestReadme": tests.TestReadme,
		"TestSpecs":  tests.TestSpecs,
	}

	var names []string
	for name := range testsByName {
		names = append(names, name)
	}
	sort.Strings(names)

	failed := false
	for _, name := range names {
		// Reset the working directory. Normally this is handled by
		// the go testing framework, but we run these tests manually,
		// because they require an internet connection, and we don't
		// want the build to require an internet connection.
		_, filename, _, _ := runtime.Caller(0)
		workingDir := path.Dir(filename)
		os.Chdir(workingDir)

		test := testsByName[name]
		if result := test(); result != nil {
			fmt.Printf("FAILED\t%s: %s\n", name, result.Error())
			failed = true
		} else {
			fmt.Printf("PASSED\t%s\n", name)
		}
	}

	if failed {
		os.Exit(1)
	}
	os.Exit(0)
}
