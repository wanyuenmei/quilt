package ipdef

import (
	"fmt"
	"math"
	"net"
)

var (
	// QuiltSubnet is the subnet under which quilt containers are given IP addresses.
	QuiltSubnet = net.IPNet{
		IP:   net.IPv4(10, 0, 0, 0).To4(),
		Mask: net.CIDRMask(8, 32),
	}

	// SubMask is the subnet mask that minions can assign container IPs within. It
	// resprents a /20 subnet.
	SubMask = net.CIDRMask(20, 32)

	// LabelSubnet is the subnet that is reserved for label IPs.
	LabelSubnet = net.IPNet{IP: QuiltSubnet.IP, Mask: SubMask}

	minionMaskBits, _ = SubMask.Size()
	quiltMaskBits, _  = QuiltSubnet.Mask.Size()

	// MaxMinionCount is the largest number of minions that can exist, based
	// on the number of available subnets
	MaxMinionCount = int(math.Pow(2, float64(minionMaskBits-quiltMaskBits))+0.5) - 1

	// GatewayIP is the address of the border router in the logical network.
	GatewayIP = net.IPv4(10, 0, 0, 1).To4()
)

// ToMac converts the given IP address string into an internal MAC address.
func ToMac(ipStr string) string {
	parsedIP := net.ParseIP(ipStr)
	if parsedIP == nil {
		return ""
	}

	ip := parsedIP.To4()
	return fmt.Sprintf("02:00:%02x:%02x:%02x:%02x", ip[0], ip[1], ip[2], ip[3])
}
