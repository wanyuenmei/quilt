package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const initialRole = "initial"
const subsequentRole = "subsequent"

func main() {
	host := os.Getenv("HOST")
	if host == "" {
		log.Fatal("You must specify a `HOST`")
	}

	role := os.Getenv("ROLE")
	if role == "" {
		log.Fatal("You must specify a `ROLE`")
	} else if role != initialRole && role != subsequentRole {
		log.Fatal("`ROLE` must be 'initial' or 'subsequent'")
	}
	log.Printf("Preparing %s", role)

	peersVar := os.Getenv("PEERS")
	peers := strings.Split(peersVar, ",")

	addrs := resolveHostname(host)

	if role == initialRole {
		for _, peer := range peers {
			checkConnection(fmt.Sprintf("%s:27017", peer))
		}
	}

	var wg sync.WaitGroup

	log.Printf("Starting Mongo server")
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := runServer(addrs); err != nil {
			log.Fatalf("Error running Mongo server: %s", err.Error())
		}
	}()

	if role == initialRole && len(peers) > 0 {
		checkConnection(fmt.Sprintf("%s:27017", host))

		log.Printf("Setting up replica set with peers: %+v", peers)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := setupReplicaSet(peers); err != nil {
				log.Fatalf("Error creating replica sets: %s", err.Error())
			}
		}()
	}

	wg.Wait()
}

type mongoConfig struct {
	Net         netCfg `json:"net"`
	Replication rsCfg  `json:"replication"`
}

type netCfg struct {
	IPs string `json:"bindIp"`
}

type rsCfg struct {
	ReplicaSetName string `json:"replSetName"`
}

func runServer(addrs []string) error {
	path := "/etc/mongod.conf"
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := json.Marshal(buildMongoConfig(addrs))
	if err != nil {
		return err
	}

	if _, err := f.Write(b); err != nil {
		return err
	}

	cmd := exec.Command("mongod", "--config", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func buildMongoConfig(addrs []string) mongoConfig {
	ips := append([]string{"127.0.0.1"}, addrs...)
	return mongoConfig{
		Net: netCfg{
			IPs: strings.Join(ips, ","),
		},
		Replication: rsCfg{
			ReplicaSetName: "rs0",
		},
	}
}

func setupReplicaSet(hosts []string) error {
	if err := executeMongoCommand("rs.initiate()"); err != nil {
		return err
	}

	for _, host := range hosts {
		err := executeMongoCommand(fmt.Sprintf("rs.add('%s')", host))
		if err != nil {
			return err
		}
	}

	return nil
}

func executeMongoCommand(op string) error {
	cmd := exec.Command("mongo", "--eval", op)
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

func checkConnection(address string) {
	for {
		conn, err := net.Dial("tcp", address)
		if err == nil {
			conn.Close()
			log.Printf("Resolved %s", address)
			return
		}

		log.Printf("Unable to resolve %s: %s. Retrying...", address, err.Error())
		time.Sleep(time.Second)
	}
}
