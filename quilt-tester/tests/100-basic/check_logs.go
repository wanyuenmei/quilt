package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
)

var machineRegex = regexp.MustCompile(`Machine-(\d+){(.+?), .*, PublicIP=(.*?),`)

func main() {
	machinesBytes, err := exec.Command("quilt", "machines").CombinedOutput()
	if err != nil {
		log.WithError(err).Fatal("Failed to retrieve machines")
	}

	machines := strings.TrimSpace(string(machinesBytes))
	fmt.Println("`quilt machines` output:")
	fmt.Println(machines)

	failed := false
	for _, machineLine := range strings.Split(machines, "\n") {
		matches := machineRegex.FindStringSubmatch(machineLine)
		if len(matches) != 4 {
			log.WithField("machine", machineLine).
				Error("Failed to parse machine")
			failed = true
			continue
		}
		id, role, ip := matches[1], matches[2], matches[3]
		fmt.Printf("%s-%s (%s) minion logs:\n", role, id, ip)
		logsOutput, err := exec.Command("quilt", "ssh", id,
			"docker", "logs", "minion").CombinedOutput()
		if err != nil {
			log.WithError(err).Error("Unable to get minion logs")
			failed = true
			continue
		}
		outputStr := string(logsOutput)
		fmt.Println(outputStr)

		if strings.Contains(outputStr, "ERROR") ||
			strings.Contains(outputStr, "WARN") {
			failed = true
		}
	}

	if failed {
		fmt.Println("FAILED")
	} else {
		fmt.Println("PASSED")
	}
}
