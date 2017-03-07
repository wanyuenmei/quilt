package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/quilt/quilt/scripts/specs-tester/tests"
)

func main() {
	defer os.RemoveAll(tests.QuiltPath)

	testsByName := map[string]func() error{
		"TestReadme": tests.TestReadme,
		"TestSpecs":  tests.TestSpecs,
	}

	var names []string
	for name := range testsByName {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		test := testsByName[name]
		runTest(name, test)
	}
}

func runTest(name string, test func() error) {
	result := test()
	if result == nil {
		fmt.Printf("PASSED\t%s\n", name)
	} else {
		fmt.Printf("FAILED\t%s: %s\n", name, result.Error())
	}
}
