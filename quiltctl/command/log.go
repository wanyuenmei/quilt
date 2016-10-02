package command

import (
	"errors"
	"flag"
	"fmt"
	"strconv"

	"github.com/NetSys/quilt/api"
	log "github.com/Sirupsen/logrus"
)

// Log is the structure for the `quilt logs` command.
type Log struct {
	host           string
	privateKey     string
	sinceTimestamp string
	showTimestamps bool
	shouldTail     bool

	targetContainer int
	flags           *flag.FlagSet
}

func (lCmd *Log) createFlagSet() {
	flags := flag.NewFlagSet("logs", flag.ExitOnError)

	flags.StringVar(&lCmd.host, "H", api.DefaultSocket,
		"the host to query for machine information")
	flags.StringVar(&lCmd.privateKey, "i", "",
		"the private key to use to connect to the host")
	flags.StringVar(&lCmd.sinceTimestamp, "since", "", "show logs since timestamp")
	flags.BoolVar(&lCmd.shouldTail, "f", false, "follow log output")
	flags.BoolVar(&lCmd.showTimestamps, "t", false, "show timestamps")

	lCmd.flags = flags
}

// Parse parses the command line arguments for the `logs` command.
func (lCmd *Log) Parse(rawArgs []string) error {
	lCmd.createFlagSet()

	if err := lCmd.flags.Parse(rawArgs); err != nil {
		return err
	}

	parsedArgs := lCmd.flags.Args()
	if len(parsedArgs) == 0 {
		return errors.New("must specify a target container")
	}

	targetContainer, err := strconv.Atoi(parsedArgs[0])
	if err != nil {
		return fmt.Errorf("target container must be a number: %s", parsedArgs[0])
	}

	lCmd.targetContainer = targetContainer
	return nil
}

// Run finds the target continer and outputs logs.
func (lCmd *Log) Run() int {
	localClient, leaderClient, err := getClients(lCmd.host)
	if err != nil {
		log.Error(err)
		return 1
	}
	defer localClient.Close()
	defer leaderClient.Close()

	containerHost, err := getContainerHost(localClient, leaderClient,
		lCmd.targetContainer)
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

	container, err := getContainer(containerClient, lCmd.targetContainer)
	if err != nil {
		log.WithError(err).Error("Error retrieving the container information " +
			"from the container host.")
		return 1
	}

	// -t allows the docker command to receive input over SSH.
	sshArgs := []string{"-t"}
	if lCmd.privateKey != "" {
		sshArgs = append(sshArgs, "-i", lCmd.privateKey)
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

	sshArgs = append(sshArgs, dockerCmd)
	if err := ssh(containerHost, sshArgs).Run(); err != nil {
		log.WithError(err).Error("Error running the logs command.")
		return 1
	}

	return 0
}

// Usage prints command usage info.
func (lCmd *Log) Usage() {
	fmt.Println("usage: quilt logs [-H=<daemon_host>] [-i=<private_key>] " +
		"<stitch_id> <command>")
	fmt.Println("`logs` fetches the logs of a container. " +
		"The container is identified by the stitch ID provided by " +
		"`quilt containers`.")
	fmt.Println("For example, to get the logs of container 5 with a " +
		"specific private key: `quilt logs -i ~/.ssh/quilt 5`")
	lCmd.flags.PrintDefaults()
}
