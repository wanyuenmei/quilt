package command

import (
	"errors"
	"flag"
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/NetSys/quilt/api/client"
	"github.com/NetSys/quilt/api/client/getter"
	"github.com/NetSys/quilt/api/util"
	"github.com/NetSys/quilt/quiltctl/ssh"
)

var (
	loginUsage = `usage: quilt login [-H=<daemon_host>] [-i=<private_key>] ` +
		`<stitch_id>
Opens a login shell in a remote docker container.

For example, to enter a shell in container 5 with a specific private key:
quilt login -i ~/.ssh/quilt 5`
)

// Login contains the options for opening shells in containers.
type Login struct {
	privateKey      string
	targetContainer string

	common *commonFlags

	clientGetter client.Getter
	sshGetter    ssh.Getter
}

// NewLoginCommand creates a new Login command instance.
func NewLoginCommand() *Login {
	return &Login{
		clientGetter: getter.New(),
		common:       &commonFlags{},
		sshGetter:    ssh.New,
	}
}

// InstallFlags sets up parsing for command line flags.
func (lCmd *Login) InstallFlags(flags *flag.FlagSet) {
	lCmd.common.InstallFlags(flags)

	flags.StringVar(&lCmd.privateKey, "i", "",
		"the private key to use to connect to the host")

	flags.Usage = func() {
		fmt.Println(loginUsage)
		flags.PrintDefaults()
	}
}

// Parse parses the command line arguments for the login command.
func (lCmd *Login) Parse(args []string) error {
	if len(args) < 1 {
		return errors.New("must specify a target container")
	}

	lCmd.targetContainer = args[0]
	return nil
}

// Run logs into a shell on the remote container.
func (lCmd *Login) Run() int {
	if !isTerminal() {
		log.Error("Cannot allocate pseudo-terminal without a terminal")
		return 1
	}

	localClient, err := lCmd.clientGetter.Client(lCmd.common.host)
	if err != nil {
		log.Error(err)
		return 1
	}
	defer localClient.Close()

	containerClient, err := lCmd.clientGetter.ContainerClient(
		localClient, lCmd.targetContainer)
	if err != nil {
		log.WithError(err).Error("Error getting container client")
		return 1
	}
	defer containerClient.Close()

	container, err := util.GetContainer(containerClient, lCmd.targetContainer)
	if err != nil {
		log.WithError(err).Error("Error getting container information")
		return 1
	}

	sshClient, err := lCmd.sshGetter(containerClient.Host(), lCmd.privateKey)
	if err != nil {
		log.WithError(err).Error("Error opening SSH connection")
		return 1
	}
	defer sshClient.Close()

	return execHelper(sshClient, container.DockerID, "sh", true)
}
