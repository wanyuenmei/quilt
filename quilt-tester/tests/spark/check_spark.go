package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func main() {
	containers, err := exec.Command("quilt", "containers").Output()
	if err != nil {
		log.WithError(err).Fatal("Unable to get containers.")
	}

	fmt.Println("`quilt containers` output:")
	fmt.Println(string(containers))

	matches := regexp.MustCompile(`.* run master.* StitchID: (\d+)`).
		FindStringSubmatch(string(containers))
	if len(matches) != 2 {
		log.Fatal("Unable to find StitchID of Spark master.")
	}

	id := matches[1]
	logs, err := exec.Command("quilt", "logs", id).CombinedOutput()
	if err != nil {
		log.WithError(err).Fatal("Unable to get Spark master logs.")
	}

	fmt.Printf("`quilt logs %s` output:\n", id)
	fmt.Println(string(logs))

	if !strings.Contains(string(logs), "Pi is roughly") {
		fmt.Println("FAILED, sparkPI did not execute correctly.")
	} else {
		fmt.Println("PASSED")
	}
}
