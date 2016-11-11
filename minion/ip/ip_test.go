package ip

import (
	"net"
	"reflect"
	"testing"
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

func TestRandomIP(t *testing.T) {
	prefix := net.IPv4(0xaa, 0xbb, 0xcc, 0xdd)
	mask := net.CIDRMask(20, 32)
	conflicts := map[string]struct{}{}

	// Only 4k IPs, in 0xfff00000. Guaranteed a collision
	for i := 0; i < 5000; i++ {
		ip := Random(conflicts, prefix, mask)
		if ip.Equal(net.IPv4zero) {
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

	if len(conflicts) < 2500 || len(conflicts) > 4096 {
		// If the code's working, this is possible but *extremely* unlikely.
		// Probably a bug.
		t.Errorf("Too few conflicts: %d", len(conflicts))
	}
}

func eq(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}
