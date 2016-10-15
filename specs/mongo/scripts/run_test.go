package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildMongoConfig(t *testing.T) {
	ip := []string{"0.0.0.0", "127.0.0.1", "192.168.52.7"}
	assert.EqualValues(t, buildMongoConfig(ip), mongoConfig{
		Net:         netCfg{IPs: "127.0.0.1,0.0.0.0,127.0.0.1,192.168.52.7"},
		Replication: rsCfg{ReplicaSetName: "rs0"},
	})
}
