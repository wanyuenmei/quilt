package quiltctl

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCommand struct {
	flagArg string
	posArg  string

	t *testing.T
}

const (
	expFlagArg = "flag"
	expPosArg  = "pos"
)

func (tCmd *testCommand) Parse(args []string) error {
	tCmd.posArg = args[0]
	return nil
}

func (tCmd *testCommand) InstallFlags(flags *flag.FlagSet) {
	flags.StringVar(&tCmd.flagArg, "arg", "", "the test arg")
}

func (tCmd *testCommand) BeforeRun() error {
	return nil
}

func (tCmd *testCommand) AfterRun() error {
	return nil
}

func (tCmd *testCommand) Run() int {
	assert.Equal(tCmd.t, expFlagArg, tCmd.flagArg)
	assert.Equal(tCmd.t, expPosArg, tCmd.posArg)
	return 0
}

func TestArgumentParsing(t *testing.T) {
	t.Parallel()

	subcommand, err := parseSubcommand("test", &testCommand{t: t},
		[]string{"-arg", expFlagArg, expPosArg})
	assert.NoError(t, err)

	subcommand.Run()
}
