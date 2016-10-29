package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildHAProxyConfig(t *testing.T) {
	ip := []string{"0.0.0.0", "127.0.0.1", "192.168.52.7"}
	assert.Equal(t, `    server 0 0.0.0.0 check
    server 1 127.0.0.1 check
    server 2 192.168.52.7 check`, buildHAProxyConfig(ip))

	assert.Equal(t, "", buildHAProxyConfig([]string{}))
}
