package command

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/NetSys/quilt/quiltctl/ssh"
	log "github.com/Sirupsen/logrus"

	"github.com/NetSys/quilt/api"
)

// Exec contains the options for running commands in containers.
type Exec struct {
	privateKey      string
	targetContainer int
	command         string

	common *commonFlags

	SSHClient ssh.Client
}

// NewExecCommand creates a new Exec command instance.
func NewExecCommand(c ssh.Client) *Exec {
	return &Exec{
		common:    &commonFlags{},
		SSHClient: c,
	}
}

// InstallFlags sets up parsing for command line flags.
func (eCmd *Exec) InstallFlags(flags *flag.FlagSet) {
	eCmd.common.InstallFlags(flags)

	flags.StringVar(&eCmd.privateKey, "i", "",
		"the private key to use to connect to the host")

	flags.Usage = func() {
		fmt.Println("usage: quilt exec [-H=<daemon_host>] [-i=<private_key>] " +
			"<stitch_id> <command>")
		fmt.Println("`exec` runs a command within the specified container. " +
			"The container is identified by the stitch ID produced by " +
			"`quilt containers`.")
		fmt.Println("For example, to get a shell in container 5 with a " +
			"specific private key: quilt exec -i ~/.ssh/quilt 5 sh")
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
	localClient, leaderClient, err := getClients(eCmd.common.host)
	if err != nil {
		log.Error(err)
		return 1
	}
	defer localClient.Close()
	defer leaderClient.Close()

	containerHost, err := getContainerHost(localClient, leaderClient,
		eCmd.targetContainer)
	if err != nil {
		log.WithError(err).
			Error("Error getting the host on which the container is running.")
		return 1
	}

	containerClient, err := getClient(api.RemoteAddress(containerHost))
	if err != nil {
		log.WithError(err).Error("Error connecting to container client.")
		return 1
	}
	defer containerClient.Close()

	container, err := getContainer(containerClient, eCmd.targetContainer)
	if err != nil {
		log.WithError(err).Error("Error retrieving the container information " +
			"from the container host.")
		return 1
	}

	if err = eCmd.SSHClient.Connect(containerHost, eCmd.privateKey); err != nil {
		log.WithError(err).Info("Error opening SSH connection")
		return 1
	}
	defer eCmd.SSHClient.Disconnect()

	if err = eCmd.SSHClient.RequestPTY(); err != nil {
		log.WithError(err).Info("Error requesting pseudo-terminal")
		return 1
	}

	command := fmt.Sprintf("docker exec -it %s %s", container.DockerID, eCmd.command)
	if err = eCmd.SSHClient.Run(command); err != nil {
		log.WithError(err).Info("Error running command over SSH")
		return 1
	}

	return 0
}
