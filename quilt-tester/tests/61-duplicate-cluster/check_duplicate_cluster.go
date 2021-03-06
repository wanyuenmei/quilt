package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/quilt/quilt/api"
	"github.com/quilt/quilt/api/client"
)

var connectionRegex = regexp.MustCompile(`Registering worker (\d+\.\d+\.\d+\.\d+:\d+)`)

func main() {
	clnt, err := client.New(api.DefaultSocket)
	if err != nil {
		log.WithError(err).Fatal("FAILED, couldn't get quiltctl client.")
	}
	defer clnt.Close()

	containers, err := clnt.QueryContainers()
	if err != nil {
		log.WithError(err).Fatal("FAILED, couldn't query containers.")
	}

	psPretty, err := exec.Command("quilt", "ps").Output()
	if err != nil {
		log.WithError(err).Fatal("FAILED, `quilt ps` failed.")
	}
	fmt.Println("`quilt ps` output:")
	fmt.Println(string(psPretty))

	var masters []string
	totalWorkers := 0
	for _, dbc := range containers {
		if strings.Join(dbc.Command, " ") == "run master" {
			id := dbc.StitchID
			masters = append(masters, id)
		} else {
			totalWorkers++
		}
	}
	if len(masters) != 2 {
		log.WithField("masters", masters).Fatal(
			"FAILED, expected 2 Spark masters.")
	}

	failed := false
	for _, master := range masters {
		logs, err := exec.Command("quilt", "logs", master).CombinedOutput()
		if err != nil {
			log.WithError(err).Fatal(
				"FAILED, unable to get Spark master logs.")
		}

		// Each cluster's workers should connect only to its own master.
		logsStr := string(logs)
		workerSet := map[string]struct{}{}
		connectionMatches := connectionRegex.FindAllStringSubmatch(logsStr, -1)
		for _, wkMatch := range connectionMatches {
			workerSet[wkMatch[1]] = struct{}{}
		}
		if workerCount := len(workerSet); workerCount != totalWorkers/2 {
			failed = true
			log.WithFields(log.Fields{
				"master":                master,
				"worker count":          workerCount,
				"expected worker count": totalWorkers / 2,
			}).Error("FAILED, wrong number of workers connected to master")
		}

		fmt.Printf("`quilt logs %s` output:\n", master)
		fmt.Println(logsStr)
	}

	if !failed {
		fmt.Println("PASSED")
	}
}
