package ip

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
)

var (
	// Rand32 stores rand.Uint32 in a variable so that it can be easily mocked out
	// for unit tests. Nondeterminism is hard to test.
	Rand32 = rand.Uint32

	// QuiltPrefix is the subnet under which quilt containers are given IP addresses
	QuiltPrefix = net.IPv4(10, 0, 0, 0)

	// QuiltMask is the subnet mask that corresponds with the Quilt subnet prefix
	QuiltMask = net.CIDRMask(8, 32)

	// SubMask is the subnet mask that minions can assign container IPs within. It
	// resprents a /20 subnet.
	SubMask = net.CIDRMask(20, 32)

	// MinionCIDRSuffix is the CIDR suffic for the minion subnet.
	MinionCIDRSuffix = "/20"

	// LabelPrefix is the subnet that is reserved for label IPs. It represents
	// 10.0.0.0/20
	LabelPrefix = net.IPv4(10, 0, 0, 0) // Labels get their own /20

	minionMaskBits, _ = SubMask.Size()
	quiltMaskBits, _  = QuiltMask.Size()

	// MaxMinionCount is the largest number of minions that can exist, based
	// on the number of available subnets
	MaxMinionCount = int(math.Pow(2, float64(minionMaskBits-quiltMaskBits))+0.5) - 1

	// GatewayIP is the address of the border router in the logical network.
	GatewayIP = net.IPv4(10, 0, 0, 1)
)

// Pool represents a set of IP addresses that can be given out to requestors.
type Pool struct {
	net.IPNet
	ipSet map[string]struct{}
}

// NewPool creates a Pool that gives out IP addresses within the given CIDR subnet.
func NewPool(prefix net.IP, mask net.IPMask) Pool {
	return Pool{
		IPNet: net.IPNet{IP: prefix, Mask: mask},
		ipSet: map[string]struct{}{},
	}
}

// AddIP adds the given IP to this Pool's IP set.
// It is an error to attempt to add an IP outside the Pool's subnet.
func (p Pool) AddIP(ip string) error {
	if addr := net.ParseIP(ip); !p.Contains(addr) {
		return fmt.Errorf("warning: IP (%s) not in subnet (%s)", ip, p)
	}
	p.ipSet[ip] = struct{}{}
	return nil
}

// Allocate selects a random IP address from the pool that hasn't already been given out.
func (p Pool) Allocate() (net.IP, error) {
	prefix := ToInt(p.IP)
	mask := MaskToInt(p.Mask)

	randStart := Rand32() & ^mask
	for offset := uint32(0); offset <= ^mask; offset++ {
		randIP := (randStart + offset) & ^mask
		ip32 := randIP | (prefix & mask)
		ip := FromInt(ip32)
		if _, ok := p.ipSet[ip.String()]; !ok {
			p.ipSet[ip.String()] = struct{}{}
			return ip, nil
		}
	}

	return nil, errors.New("IP pool exhausted")
}

// MaskToInt takes in a CIDR Mask and return the integer representation of it.
func MaskToInt(mask net.IPMask) uint32 {
	bits, _ := mask.Size()
	return 0xffffffff ^ uint32(0xffffffff>>uint(bits))
}

// FromInt converts the given integer into the equivalent IP address.
func FromInt(ip32 uint32) net.IP {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, ip32)
	return net.IP(b)
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

// ToInt converts the given IP address into the equivalent integer representation.
func ToInt(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip.To4())
}
