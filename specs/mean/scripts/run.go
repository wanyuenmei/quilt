package main

import (
	"log"
	"os"
	"os/exec"
	"time"

	"gopkg.in/mgo.v2"

	"github.com/NetSys/quilt/specs"
)

func main() {
	dialInfo, err := mgo.ParseURL(os.Getenv("MONGO_URI"))
	if err != nil {
		log.Fatalf("Failed to parse MONGO_URI: %s\n", err)
	}

	// Due to a cache in the go name resolution code, we shell out to ping first
	// before relying on Dial.  Otherwise, dial would cache /etc/hosts before it was
	// ready causing irrecoverable problems.
	if err := specs.PingWait(dialInfo.Addrs); err != nil {
		log.Fatalf("Error ping wait: %s", err.Error())
	}

	dialInfo.Timeout = 5 * time.Second
	for {
		if _, err = mgo.DialWithInfo(dialInfo); err == nil {
			log.Print("Connected to Mongo")
			break
		}

		log.Printf("Failed to connect: %s\n", err.Error())
		time.Sleep(5 * time.Second)
	}

	cmd := exec.Command("npm", "start")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error running server: %s", err.Error())
	}
}
