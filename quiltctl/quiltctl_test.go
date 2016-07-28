package quiltctl

import (
	"flag"
	"testing"

	"github.com/NetSys/quilt/quiltctl/command"
)

type testCommand struct {
	arg string

	t *testing.T
}

const expArgVal = "val"

func (tCmd *testCommand) Parse(args []string) error {
	flags := flag.NewFlagSet("test", flag.ExitOnError)

	flags.StringVar(&tCmd.arg, "arg", "", "the test arg")

	return flags.Parse(args)
}

func (tCmd *testCommand) Run() int {
	if tCmd.arg != expArgVal {
		tCmd.t.Errorf("Bad argument value for testCommand: expected %s, got %s",
			expArgVal, tCmd.arg)
	}
	return 0
}

func (tCmd *testCommand) Usage() {
}

func TestArgumentParsing(t *testing.T) {
	t.Parallel()

	commands = map[string]command.SubCommand{
		"testCommand": &testCommand{t: t},
	}
	subcommand, err := parseSubcommand("testCommand", []string{"-arg", expArgVal})
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
