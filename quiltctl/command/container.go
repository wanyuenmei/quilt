package command

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/NetSys/quilt/db"
)

// Container contains the options for querying containers.
type Container struct {
	*commonFlags
}

// NewContainerCommand creates a new Container command instance.
func NewContainerCommand() *Container {
	return &Container{
		commonFlags: &commonFlags{},
	}
}

// Parse parses the command line arguments for the container command.
func (cCmd *Container) Parse(args []string) error {
	return nil
}

// Run retrieves and prints the requested containers.
func (cCmd *Container) Run() int {
	localClient, err := getClient(cCmd.host)
	if err != nil {
		log.Error(err)
		return 1
	}

	c, err := getLeaderClient(localClient)
	localClient.Close()
	if err != nil {
		log.WithError(err).Error("Error connecting to leader.")
		return 1
	}
	defer c.Close()

	containers, err := c.QueryContainers()
	if err != nil {
		log.WithError(err).Error("Unable to query containers.")
		return 1
	}

	str := containersStr(containers)
	fmt.Print(str)

	return 0
}

func containersStr(containers []db.Container) string {
	var containersStr string
	for _, c := range containers {
		containersStr += fmt.Sprintf("%v\n", c)
	}

	return containersStr
}
