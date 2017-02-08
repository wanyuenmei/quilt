package command

import (
	"errors"
	"flag"
	"fmt"

	"github.com/NetSys/quilt/api/client"
	"github.com/NetSys/quilt/api/client/getter"
	"github.com/NetSys/quilt/api/util"
	"github.com/NetSys/quilt/quiltctl/ssh"

	log "github.com/Sirupsen/logrus"
)

// Log is the structure for the `quilt logs` command.
type Log struct {
	privateKey     string
	sinceTimestamp string
	showTimestamps bool
	shouldTail     bool

	targetContainer string

	sshGetter    ssh.Getter
	clientGetter client.Getter
	common       *commonFlags
}

// NewLogCommand creates a new Log command instance.
func NewLogCommand() *Log {
	return &Log{
		sshGetter:    ssh.New,
		clientGetter: getter.New(),
		common:       &commonFlags{},
	}
}

// InstallFlags sets up parsing for command line flags.
func (lCmd *Log) InstallFlags(flags *flag.FlagSet) {
	lCmd.common.InstallFlags(flags)

	flags.StringVar(&lCmd.privateKey, "i", "",
		"the private key to use to connect to the host")
	flags.StringVar(&lCmd.sinceTimestamp, "since", "", "show logs since timestamp")
	flags.BoolVar(&lCmd.shouldTail, "f", false, "follow log output")
	flags.BoolVar(&lCmd.showTimestamps, "t", false, "show timestamps")

	flags.Usage = func() {
		fmt.Println("usage: quilt logs [-H=<daemon_host>] [-i=<private_key>] " +
			"<stitch_id>")
		fmt.Println("`logs` fetches the logs of a container. " +
			"The container is identified by the stitch ID provided by " +
			"`quilt containers`.")
		fmt.Println("For example, to get the logs of container 5 with a " +
			"specific private key: `quilt logs -i ~/.ssh/quilt 5`")
		flags.PrintDefaults()
	}
}

// Parse parses the command line arguments for the `logs` command.
func (lCmd *Log) Parse(args []string) error {
	if len(args) == 0 {
		return errors.New("must specify a target container")
	}

	lCmd.targetContainer = args[0]
	return nil
}

// Run finds the target continer and outputs logs.
func (lCmd *Log) Run() int {
	localClient, err := lCmd.clientGetter.Client(lCmd.common.host)
	if err != nil {
		log.Error(err)
		return 1
	}
	defer localClient.Close()

	containerClient, err := lCmd.clientGetter.ContainerClient(
		localClient, lCmd.targetContainer)
	if err != nil {
		log.WithError(err).Error("Error getting container client.")
		return 1
	}

	container, err := util.GetContainer(containerClient, lCmd.targetContainer)
	if err != nil {
		log.WithError(err).Error("Error getting container information.")
		return 1
	}

	dockerCmd := "docker logs"
	if lCmd.sinceTimestamp != "" {
		dockerCmd += fmt.Sprintf(" --since=%s", lCmd.sinceTimestamp)
	}
	if lCmd.showTimestamps {
		dockerCmd += " --timestamps"
	}
	if lCmd.shouldTail {
		dockerCmd += " --follow"
	}
	dockerCmd += " " + container.DockerID

	sshClient, err := lCmd.sshGetter(containerClient.Host(), lCmd.privateKey)
	if err != nil {
		log.WithError(err).Info("Error opening SSH connection")
		return 1
	}
	defer sshClient.Close()

	if err = sshClient.Run(false, dockerCmd); err != nil {
		log.WithError(err).Info("Error running command over SSH")
		return 1
	}

	return 0
}
