package command

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	log "github.com/Sirupsen/logrus"

	"github.com/NetSys/quilt/api/client"
	"github.com/NetSys/quilt/api/client/getter"
)

// SSH contains the options for SSHing into machines.
type SSH struct {
	targetMachine int
	sshArgs       []string

	common       *commonFlags
	clientGetter client.Getter
}

// NewSSHCommand creates a new SSH command instance.
func NewSSHCommand() *SSH {
	return &SSH{
		clientGetter: getter.New(),
		common:       &commonFlags{},
	}
}

// InstallFlags sets up parsing for command line flags.
func (sCmd *SSH) InstallFlags(flags *flag.FlagSet) {
	sCmd.common.InstallFlags(flags)

	flags.Usage = func() {
		fmt.Println("usage: quilt ssh [-H=<daemon_host>] <machine_num> " +
			"[ssh_options]")
		fmt.Println("`ssh` creates a SSH session to the specified machine. " +
			"The machine is identified the database ID produced by " +
			"`quilt queryMachines`.")
		fmt.Println("For example, to SSH to machine 5 with a specific " +
			"private key: quilt ssh 5 -i ~/.ssh/quilt")
		flags.PrintDefaults()
	}
}

// Parse parses the command line arguments for the ssh command.
func (sCmd *SSH) Parse(args []string) error {
	if len(args) == 0 {
		return errors.New("must specify a target machine")
	}

	targetMachine, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("target machine must be a number: %s", args[0])
	}

	sCmd.targetMachine = targetMachine
	sCmd.sshArgs = args[1:]
	return nil
}

// Run SSHs into the given machine.
func (sCmd *SSH) Run() int {
	c, err := sCmd.clientGetter.Client(sCmd.common.host)
	if err != nil {
		log.Error(err)
		return 1
	}
	defer c.Close()

	machines, err := c.QueryMachines()
	if err != nil {
		log.WithError(err).Error("Unable to query machines.")
		return 1
	}

	var host string
	for _, m := range machines {
		if m.ID == sCmd.targetMachine {
			host = m.PublicIP
			break
		}
	}

	if host == "" {
		missingMachineMsg :=
			fmt.Sprintf("Unable to find machine `%d`.\n", sCmd.targetMachine)
		missingMachineMsg += "Available machines:\n"
		for _, m := range machines {
			missingMachineMsg += fmt.Sprintf("%v\n", m)
		}
		log.Error(missingMachineMsg)
		return 1
	}

	if err = runSSHCommand(host, sCmd.sshArgs).Run(); err != nil {
		log.WithError(err).Error("Error executing the SSH command")
		return 1
	}
	return 0
}

// Stored in a variable so we can mock it out for unit tests.
var runSSHCommand = func(host string, args []string) *exec.Cmd {
	baseArgs := []string{fmt.Sprintf("quilt@%s", host),
		"-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null"}

	cmd := exec.Command("ssh", append(baseArgs, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}
