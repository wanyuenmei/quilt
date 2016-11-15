package main

import (
	"log"
	"net"
	"os"
	"os/exec"
	"time"

	"gopkg.in/mgo.v2"
)

func main() {
	host := os.Getenv("HOST")
	if host == "" {
		log.Fatal("You must specify a `HOST`")
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("You must specify a `MONGO_URI`")
	}

	resolveHostname(host)
	waitForMongo(mongoURI)

	if err := runServer(); err != nil {
		log.Fatalf("Error running server: %s", err.Error())
	}
}

func runServer() error {
	cmd := exec.Command("npm", "start")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func resolveHostname(hostname string) []string {
	for {
		addrs, err := net.LookupHost(hostname)
		if err == nil {
			log.Printf("Resolved %s to %+v", hostname, addrs)
			return addrs
		}

		log.Printf("Unable to resolve %s: %s. Retrying...", hostname, err.Error())
		time.Sleep(time.Second)
	}
}

func waitForMongo(mongoURI string) {
	dialInfo, err := mgo.ParseURL(mongoURI)
	if err != nil {
		log.Fatalf("Failed to parse Mongo URI: %s\n", err.Error())
	}

	for {
		_, err = mgo.DialWithInfo(dialInfo)
		if err == nil {
			log.Print("Connected to Mongo")
			return
		}

		log.Printf("Failed to connect: %s\n", err.Error())
		time.Sleep(time.Second)
	}
}
