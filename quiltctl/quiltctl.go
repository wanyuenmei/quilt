package quiltctl

import (
	"flag"
	"fmt"
	"os"

	"github.com/NetSys/quilt/quiltctl/command"
	"github.com/NetSys/quilt/quiltctl/ssh"

	log "github.com/Sirupsen/logrus"
)

var commands = map[string]command.SubCommand{
	"containers": command.NewContainerCommand(),
	"daemon":     command.NewDaemonCommand(),
	"exec":       command.NewExecCommand(ssh.NewNativeClient()),
	"get":        &command.Get{},
	"inspect":    &command.Inspect{},
	"logs":       command.NewLogCommand(ssh.NewNativeClient()),
	"machines":   command.NewMachineCommand(),
	"minion":     &command.Minion{},
	"run":        command.NewRunCommand(),
	"ssh":        command.NewSSHCommand(),
	"stop":       command.NewStopCommand(),
}

// Run parses and runs the quiltctl subcommand given the command line arguments.
func Run(args []string) {
	if len(args) == 0 {
		usage()
	}

	cmd, err := parseSubcommand(args[0], args[1:])
	if err != nil {
		log.WithError(err).Error("Unable to parse subcommand.")
		usage()
	}

	os.Exit(cmd.Run())
}

// HasSubcommand returns true if quiltctl has a subcommand for the given name.
func HasSubcommand(name string) bool {
	_, ok := commands[name]
	return ok
}

func parseSubcommand(name string, args []string) (command.SubCommand, error) {
	if !HasSubcommand(name) {
		return nil, fmt.Errorf("unrecognized subcommand: %s", name)
	}

	cmd := commands[name]
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	cmd.InstallFlags(flags)
	if err := flags.Parse(args); err != nil {
		flags.Usage()
		return nil, err
	}

	if err := cmd.Parse(flags.Args()); err != nil {
		flags.Usage()
		return nil, err
	}

	return cmd, nil
}

func usage() {
	flag.Usage()
	os.Exit(1)
}
