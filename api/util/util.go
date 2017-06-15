package util

import (
	"fmt"

	"github.com/quilt/quilt/db"
)

// GetContainer retrieves the container with the given stitchID.
func GetContainer(containers []db.Container, stitchID string) (db.Container, error) {
	var choice *db.Container
	for _, c := range containers {
		if len(stitchID) > len(c.StitchID) ||
			c.StitchID[0:len(stitchID)] != stitchID {
			continue
		}

		if choice != nil {
			err := fmt.Errorf("ambiguous stitchIDs %s and %s",
				choice.StitchID, c.StitchID)
			return db.Container{}, err
		}

		copy := c
		choice = &copy
	}

	if choice != nil {
		return *choice, nil
	}

	return db.Container{}, fmt.Errorf("no container with stitchID %q", stitchID)
}

// GetPublicIP returns the public IP for the machine with the given private IP.
func GetPublicIP(machines []db.Machine, privateIP string) (string, error) {
	for _, m := range machines {
		if m.PrivateIP == privateIP {
			return m.PublicIP, nil
		}
	}

	return "", fmt.Errorf("no machine with private IP %s", privateIP)
}
