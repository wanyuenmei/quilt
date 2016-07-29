package command

import (
	"github.com/NetSys/quilt/inspect"
)

// Inspect contains the options for inspecting Stitches.
type Inspect struct {
	args []string
}

// Parse parses the command line arguments for the inspect command.
func (iCmd *Inspect) Parse(args []string) error {
	iCmd.args = args
	return nil
}

// Run inspects the provided Stitch.
func (iCmd *Inspect) Run() int {
	return inspect.Main(iCmd.args)
}

// Usage prints the usage for the inspect command.
func (iCmd *Inspect) Usage() {
	inspect.Usage()
}
