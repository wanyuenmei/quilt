package ip

import (
	"math/rand"
	"net"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestSyncIPs(t *testing.T) {
	prefix := net.IPv4(10, 0, 0, 0)

	nextRand := uint32(0)
	Rand32 = func() uint32 {
		ret := nextRand
		nextRand++
		return ret
	}

	defer func() {
		Rand32 = rand.Uint32
	}()

	ipMap := map[string]string{
		"a": "",
		"b": "",
		"c": "",
	}

	Sync(ipMap, prefix)

	// 10.0.0.1 is reserved for the default gateway
	exp := sliceToSet([]string{"10.0.0.0", "10.0.0.2", "10.0.0.3"})
	ipSet := map[string]struct{}{}
	for _, ip := range ipMap {
		ipSet[ip] = struct{}{}
	}

	if !eq(ipSet, exp) {
		t.Error(spew.Sprintf("Unexpected IP allocations."+
			"\nFound %s\nExpected %s\nMap %s",
			ipSet, exp, ipMap))
	}

	ipMap["a"] = "junk"

	Sync(ipMap, prefix)

	aIP := ipMap["a"]
	expected := "10.0.0.4"
	if aIP != expected {
		t.Error(spew.Sprintf("Unexpected IP allocations.\nFound %s\nExpected %s",
			aIP, expected))
	}

	// Force collisions
	Rand32 = func() uint32 {
		return 4
	}

	ipMap["b"] = "junk"

	Sync(ipMap, prefix)

	if ip, _ := ipMap["b"]; ip != "" {
		t.Error(spew.Sprintf("Expected IP deletion, found %s", ip))
	}
}

func TestParseIP(t *testing.T) {
	res := Parse("1.0.0.0", 0x01000000, 0xff000000)
	if res != 0x01000000 {
		t.Errorf("Parse expected 0x%x, got 0x%x", 0x01000000, res)
	}

	res = Parse("2.0.0.1", 0x01000000, 0xff000000)
	if res != 0 {
		t.Errorf("Parse expected 0x%x, got 0x%x", 0, res)
	}

	res = Parse("a", 0x01000000, 0xff000000)
	if res != 0 {
		t.Errorf("Parse expected 0x%x, got 0x%x", 0, res)
	}
}

func TestRandomIP(t *testing.T) {
	prefix := uint32(0xaabbccdd)
	mask := uint32(0xfffff000)

	conflicts := map[uint32]struct{}{}

	// Only 4k IPs, in 0xfff00000. Guaranteed a collision
	for i := 0; i < 5000; i++ {
		ip := randomIP(conflicts, prefix, mask)
		if ip == 0 {
			continue
		}

		if _, ok := conflicts[ip]; ok {
			t.Fatalf("IP Double allocation: 0x%x", ip)
		}

		if prefix&mask != ip&mask {
			t.Fatalf("Bad IP allocation: 0x%x & 0x%x != 0x%x",
				ip, mask, prefix&mask)
		}

		conflicts[ip] = struct{}{}
	}

	if len(conflicts) < 2500 || len(conflicts) > 4096 {
		// If the code's working, this is possible but *extremely* unlikely.
		// Probably a bug.
		t.Errorf("Too few conflicts: %d", len(conflicts))
	}
}

func sliceToSet(slice []string) map[string]struct{} {
	res := map[string]struct{}{}
	for _, s := range slice {
		res[s] = struct{}{}
	}
	return res
}

func eq(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}
