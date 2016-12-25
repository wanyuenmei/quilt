package specs

import (
	"log"
	"os/exec"
	"strings"
	"time"
)

// PingWait blocks until all `addrs` are pingable.
func PingWait(addrs []string) error {
	var err error
	for _, fullAddr := range addrs {
		addr := strings.Split(fullAddr, ":")[0]
		log.Printf("Pinging %s", addr)
		for pinged := false; !pinged; pinged, err = ping(addr) {
			if err != nil {
				return err
			}
			time.Sleep(time.Second)
		}
		log.Printf("Successfully pinged %s", addr)
	}
	return nil
}

func ping(addr string) (bool, error) {
	err := exec.Command("ping", "-c1", addr).Run()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
