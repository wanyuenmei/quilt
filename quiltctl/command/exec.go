package command

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/NetSys/quilt/api/client"
	"github.com/NetSys/quilt/api/client/getter"
	"github.com/NetSys/quilt/api/util"
	"github.com/NetSys/quilt/quiltctl/ssh"
)

// Exec contains the options for running commands in containers.
type Exec struct {
	privateKey      string
	targetContainer int
	command         string
	allocatePTY     bool

	common *commonFlags

	SSHClient    ssh.Client
	clientGetter client.Getter
}

// NewExecCommand creates a new Exec command instance.
func NewExecCommand(c ssh.Client) *Exec {
	return &Exec{
		common:       &commonFlags{},
		SSHClient:    c,
		clientGetter: getter.New(),
	}
}

// InstallFlags sets up parsing for command line flags.
func (eCmd *Exec) InstallFlags(flags *flag.FlagSet) {
	eCmd.common.InstallFlags(flags)

	flags.StringVar(&eCmd.privateKey, "i", "",
		"the private key to use to connect to the host")
	flags.BoolVar(&eCmd.allocatePTY, "t", false,
		"attempt to allocate a pseudo-terminal")

	flags.Usage = func() {
		fmt.Println("usage: quilt exec [-H=<daemon_host>] [-i=<private_key>] " +
			"[-t] <stitch_id> <command>")
		fmt.Println("`exec` runs a command within the specified container. " +
			"The container is identified by the stitch ID produced by " +
			"`quilt containers`.")
		fmt.Println("For example, to get a shell in container 5 with a " +
			"specific private key: quilt exec -t -i ~/.ssh/quilt 5 sh")
		flags.PrintDefaults()
	}
}

// Parse parses the command line arguments for the exec command.
func (eCmd *Exec) Parse(args []string) error {
	if len(args) < 2 {
		return errors.New("must specify a target container and command")
	}

	targetContainer, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("target container must be a number: %s", args[0])
	}

	eCmd.targetContainer = targetContainer
	eCmd.command = strings.Join(args[1:], " ")
	return nil
}

// Run finds the target continer, and executes the given command in it.
func (eCmd *Exec) Run() int {
	if eCmd.allocatePTY && !isTerminal() {
		log.Error("Cannot allocate pseudo-terminal without a terminal")
		return 1
	}

	localClient, err := eCmd.clientGetter.Client(eCmd.common.host)
	if err != nil {
		log.Error(err)
		return 1
	}
	defer localClient.Close()

	containerClient, err := eCmd.clientGetter.ContainerClient(
		localClient, eCmd.targetContainer)
	if err != nil {
		log.WithError(err).Error("Error getting container client")
		return 1
	}

	container, err := util.GetContainer(containerClient, eCmd.targetContainer)
	if err != nil {
		log.WithError(err).Error("Error getting container information")
		return 1
	}

	err = eCmd.SSHClient.Connect(containerClient.Host(), eCmd.privateKey)
	if err != nil {
		log.WithError(err).Info("Error opening SSH connection")
		return 1
	}
	defer eCmd.SSHClient.Disconnect()

	if eCmd.allocatePTY {
		if err = eCmd.SSHClient.RequestPTY(); err != nil {
			log.WithError(err).Info("Error requesting pseudo-terminal")
			return 1
		}
	}

	var flags string
	if eCmd.allocatePTY {
		flags = "-it"
	}
	command := strings.Join(
		[]string{"docker exec", flags, container.DockerID, eCmd.command}, " ")
	if err = eCmd.SSHClient.Run(command); err != nil {
		log.WithError(err).Info("Error running command over SSH")
		return 1
	}

	return 0
}

var isTerminal = func() bool {
	return terminal.IsTerminal(int(os.Stdout.Fd()))
}
