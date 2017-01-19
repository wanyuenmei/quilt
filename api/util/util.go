package util

import (
	"fmt"

	"github.com/NetSys/quilt/api/client"
	"github.com/NetSys/quilt/db"
)

// GetContainer retrieves the container tracked by the given client with the
// given stitchID.
func GetContainer(c client.Client, stitchID string) (db.Container, error) {
	containers, err := c.QueryContainers()
	if err != nil {
		return db.Container{}, err
	}

	for _, c := range containers {
		if c.StitchID == stitchID {
			return c, nil
		}
	}

	return db.Container{}, fmt.Errorf("no container with stitchID %q", stitchID)
}
