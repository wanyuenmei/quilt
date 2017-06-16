package ipdef

import (
	"fmt"
	"net"
	"syscall"
)

var (
	// QuiltSubnet is the subnet under which Quilt containers and load balancers
	// are given IP addresses.
	QuiltSubnet = net.IPNet{
		IP:   net.IPv4(10, 0, 0, 0),
		Mask: net.CIDRMask(8, 32),
	}

	// GatewayIP is the address of the border router in the logical network.
	GatewayIP = net.IPv4(10, 0, 0, 1)

	// GatewayMac is the Mac address of the default gateway.
	GatewayMac = IPToMac(GatewayIP)

	// LoadBalancerIP is the IP address used to generate the MAC address of the
	// load balancer router. It isn't directly used for routing, but we need to
	// ensure that nothing else allocates this IP so that the MAC does not
	// conflict.
	LoadBalancerIP = net.IPv4(10, 0, 0, 2)

	// LoadBalancerMac is the MAC address of the load balancer router.
	LoadBalancerMac = IPToMac(LoadBalancerIP)

	// QuiltBridge is the Open vSwitch bridge controlled by the Quilt minion.
	QuiltBridge = "quilt-int"

	// OvnBridge is the Open vSwitch bridge controlled by OVN.
	OvnBridge = "br-int"
)

// IPStrToMac converts the given IP address string into a MAC address.
func IPStrToMac(ipStr string) string {
	parsedIP := net.ParseIP(ipStr)
	if parsedIP == nil {
		return ""
	}
	return IPToMac(parsedIP)
}

// IPToMac converts the given IP address into a MAC address.
func IPToMac(ip net.IP) string {
	ip = ip.To4()
	return fmt.Sprintf("02:00:%02x:%02x:%02x:%02x", ip[0], ip[1], ip[2], ip[3])
}

// Allow mocking out for unit tests.
var ifNameSize = syscall.IFNAMSIZ

// IFName transforms a string into something suitable for an interface name.
func IFName(name string) string {
	// The IFNAMESIZ #define is the size of a C buffer, not the length of a string.
	// Thus, it assumes one NULL character at the end which we can't overwrite.
	size := ifNameSize - 1

	if len(name) < size {
		return name
	}
	return name[0:size]
}

// PatchPorts takes an ID and converts it to two patch port names.  One for the
// QuiltBridge and one for the OvnBridge.
func PatchPorts(id string) (br, quilt string) {
	return IFName("br_" + id), IFName("q_" + id)
}
