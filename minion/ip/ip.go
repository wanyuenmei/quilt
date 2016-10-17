package ip

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"

	"github.com/NetSys/quilt/minion/network"

	log "github.com/Sirupsen/logrus"
)

var (
	// Rand32 stores rand.Uint32 in a variable so that it can be easily mocked out
	// for unit tests. Nondeterminism is hard to test.
	Rand32 = rand.Uint32
)

// Sync takes a map of IDs to IPs and creates an IP address for every entry that's
// missing one.
func Sync(ipMap map[string]string, prefixIP net.IP) {
	prefix := binary.BigEndian.Uint32(prefixIP.To4())
	mask := uint32(0xffff0000)

	var unassigned []string
	ipSet := map[uint32]struct{}{}
	for k, ipString := range ipMap {
		ip := Parse(ipString, prefix, mask)
		if ip != 0 {
			ipSet[ip] = struct{}{}
		} else {
			unassigned = append(unassigned, k)
		}
	}

	// Don't assign the IP of the default gateway
	ipSet[Parse(network.GatewayIP, prefix, mask)] = struct{}{}
	for _, k := range unassigned {
		ip32 := randomIP(ipSet, prefix, mask)
		if ip32 == 0 {
			log.Errorf("Failed to allocate IP for %s.", k)
			ipMap[k] = ""
			continue
		}

		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, ip32)

		ipMap[k] = net.IP(b).String()
		ipSet[ip32] = struct{}{}
	}
}

// Parse takes in an IP string, and parses it with respect to the given prefix and mask.
func Parse(ipStr string, prefix, mask uint32) uint32 {
	ip := net.ParseIP(ipStr).To4()
	if ip == nil {
		return 0
	}

	ip32 := binary.BigEndian.Uint32(ip)
	if ip32&mask != prefix {
		return 0
	}

	return ip32
}

// ToMac converts the given IP address string into an internal MAC address.
func ToMac(ipStr string) string {
	parsedIP := net.ParseIP(ipStr)
	if parsedIP == nil {
		return ""
	}

	ip := parsedIP.To4()
	return fmt.Sprintf("02:00:%02x:%02x:%02x:%02x", ip[0], ip[1], ip[2], ip[3])
}

// Select a random IP address in the given prefix and mask that isn't already
// present in conflicts.
// Returns 0 on failure.
func randomIP(conflicts map[uint32]struct{}, prefix, mask uint32) uint32 {
	for i := 0; i < 256; i++ {
		ip32 := (Rand32() & ^mask) | (prefix & mask)
		if _, ok := conflicts[ip32]; !ok {
			return ip32
		}
	}

	return 0
}
