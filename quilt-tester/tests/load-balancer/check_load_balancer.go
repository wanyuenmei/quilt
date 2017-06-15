package main

import (
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/quilt/quilt/api"
	"github.com/quilt/quilt/api/client"
	"github.com/quilt/quilt/db"
)

const (
	fetcherLabel      = "fetcher"
	loadBalancedLabel = "loadBalanced"
)

func main() {
	c, err := client.New(api.DefaultSocket)
	if err != nil {
		log.WithError(err).Fatal("FAILED, couldn't get local client")
	}
	defer c.Close()

	containers, err := c.QueryContainers()
	if err != nil {
		log.WithError(err).Fatal("FAILED, couldn't get containers")
	}

	var loadBalancedContainers []db.Container
	var fetcherID string
	for _, c := range containers {
		if contains(c.Labels, fetcherLabel) {
			fetcherID = c.StitchID
		}
		if contains(c.Labels, loadBalancedLabel) {
			loadBalancedContainers = append(loadBalancedContainers, c)
		}
	}
	log.WithField("expected unique responses", len(loadBalancedContainers)).
		Info("Starting fetching..")

	if fetcherID == "" {
		log.Fatal("FAILED, couldn't find fetcher")
	}

	loadBalancedCounts := map[string]int{}
	for i := 0; i < len(loadBalancedContainers)*5; i++ {
		outBytes, err := exec.Command("quilt", "ssh", fetcherID,
			"wget", "-q", "-O", "-", loadBalancedLabel+".q").
			CombinedOutput()
		if err != nil {
			log.WithError(err).Fatal("Unable to GET")
		}

		loadBalancedCounts[strings.TrimSpace(string(outBytes))]++
	}

	log.WithField("counts", loadBalancedCounts).Info("Fetching completed")
	if len(loadBalancedCounts) < len(loadBalancedContainers) {
		log.Fatal("FAILED, some containers not load balanced")
	}
	log.Info("PASSED")
}

func contains(lst []string, key string) bool {
	for _, v := range lst {
		if v == key {
			return true
		}
	}
	return false
}
