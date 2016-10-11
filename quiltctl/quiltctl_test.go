package quiltctl

import (
	"flag"
	"testing"

	"github.com/NetSys/quilt/quiltctl/command"
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

func (tCmd *testCommand) Run() int {
	if tCmd.flagArg != expFlagArg {
		tCmd.t.Errorf("Bad argument value for testCommand: expected %s, got %s",
			expFlagArg, tCmd.flagArg)
	}
	if tCmd.posArg != expPosArg {
		tCmd.t.Errorf("Bad argument value for testCommand: expected %s, got %s",
			expPosArg, tCmd.posArg)
	}
	return 0
}

func TestArgumentParsing(t *testing.T) {
	t.Parallel()

	commands = map[string]command.SubCommand{
		"testCommand": &testCommand{t: t},
	}
	subcommand, err := parseSubcommand("testCommand",
		[]string{"-arg", expFlagArg, expPosArg})
	if err != nil {
		t.Errorf("Unexpected error: %s\n", err.Error())
		return
	}

	subcommand.Run()
}

func TestUnknownSubcommand(t *testing.T) {
	t.Parallel()

	_, err := parseSubcommand("undefinedSubcommand", []string{})
	expErr := "unrecognized subcommand: undefinedSubcommand"
	if err == nil {
		t.Error("No error returned")
		return
	}

	if err.Error() != expErr {
		t.Errorf("Expected error \"%s\", but got \"%s\".", expErr, err.Error())
		return
	}
}
