package ip

import (
	"math/rand"
	"net"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskToInt(t *testing.T) {
	mask := net.CIDRMask(16, 32)
	if MaskToInt(mask) != 0xffff0000 {
		t.Fatalf("Wrong mask int, expected 0xffff0000, got %x", MaskToInt(mask))
	}

	mask = net.CIDRMask(19, 32)
	if MaskToInt(mask) != 0xffffe000 {
		t.Fatalf("Wrong mask int, expected 0xffffe000, got %x", MaskToInt(mask))
	}

	mask = net.CIDRMask(32, 32)
	if MaskToInt(mask) != 0xffffffff {
		t.Fatalf("Wrong mask int, expected 0xffffffff, got %x", MaskToInt(mask))
	}
}

func TestAllocate(t *testing.T) {
	prefix := net.IPv4(0xab, 0xcd, 0xe0, 0x00)
	mask := net.CIDRMask(20, 32)
	pool := NewPool(prefix, mask)
	conflicts := map[string]struct{}{}

	// Only 4k IPs, in 0xfffff000. Guaranteed a collision
	for i := 0; i < 5000; i++ {
		ip, err := pool.Allocate()
		if err != nil {
			continue
		}

		if _, ok := conflicts[ip.String()]; ok {
			t.Fatalf("IP Double allocation: 0x%x", ip)
		}

		if !prefix.Mask(mask).Equal(ip.Mask(mask)) {
			t.Fatalf("Bad IP allocation: %v & %v != %v",
				ip, mask, prefix.Mask(mask))
		}

		conflicts[ip.String()] = struct{}{}
	}

	if len(conflicts) != len(pool.ipSet) {
		t.Fatalf("The IP Pool has a different number of IPs than conflicts.\n"+
			"Got: %d, Expected: %d", len(pool.ipSet), len(conflicts))
	}

	if len(conflicts) < 2500 || len(conflicts) > 4096 {
		// If the code's working, this is possible but *extremely* unlikely.
		// Probably a bug.
		t.Errorf("Too few conflicts: %d", len(conflicts))
	}
}

func TestAddIP(t *testing.T) {
	// Test that added IPs are not allocated.
	for i := 0; i < 10; i++ {
		testAddIP(t)
	}

	prefix := net.IPv4(10, 0, 0, 0)
	mask := net.CIDRMask(20, 32)
	pool := NewPool(prefix, mask)

	// Test that AddIP errors when the IP is out of the subnet.
	for i := 0; i < 256; i++ {
		a, b, c, d := 11+rand.Intn(200), rand.Intn(200),
			rand.Intn(200), rand.Intn(200)
		addr := net.IPv4(byte(a), byte(b), byte(c), byte(d))
		err := pool.AddIP(addr.String())
		assert.NotNil(t, err)
	}
}

func testAddIP(t *testing.T) {
	prefix := net.IPv4(10, 0, 0, 0)
	mask := net.CIDRMask(28, 32)
	pool := NewPool(prefix, mask)

	ipSet := map[string]struct{}{}
	for i := 0; i < 16; i++ {
		ipSet["10.0.0."+strconv.Itoa(i)] = struct{}{}
	}

	for i := 0; i < 4; i++ {
		j := ""
		for {
			j = "10.0.0." + strconv.Itoa(rand.Intn(16))
			if _, ok := ipSet[j]; ok {
				break
			}
		}

		pool.AddIP(j)
		delete(ipSet, j)
	}

	allocSet := map[string]struct{}{}
	for i := 0; i < 12; i++ {
		addr, err := pool.Allocate()
		assert.Nil(t, err)
		allocSet[addr.String()] = struct{}{}
	}

	assert.Equal(t, ipSet, allocSet)
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
