package network

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
)

// DelVeth deletes a virtual ethernet interface.
var DelVeth = func(endpointID string) error {
	_, name := VethPairNames(endpointID)
	if err := linkDelete("", name); err != nil {
		return err
	}
	return nil
}

// AddVeth creates a virtual ethernet interface.
var AddVeth = func(endpointID string) (string, error) {
	tmpPeer, name := VethPairNames(endpointID)

	// Create veth pair
	err := ipExec("", "link add %s type veth peer name %s", name, tmpPeer)
	if err != nil {
		return "", fmt.Errorf("error adding veth %s with peer %s: %v",
			name, tmpPeer, err)
	}

	// Set the host side link status to up
	if err = ipExec("", "link set %s up", name); err != nil {
		return "", fmt.Errorf("error bringing veth %s up: %v", name, err)
	}

	if err = ipExec("", "link set dev %s mtu %d", tmpPeer, innerMTU); err != nil {
		return "", fmt.Errorf("error setting peer %s MTU: %v", tmpPeer, err)
	}

	return tmpPeer, nil
}

// LinkExists reports whether or not the virtual link exists.
var LinkExists = func(namespace, name string) (bool, error) {
	cmd := fmt.Sprintf("ip link show %s", name)
	if namespace != "" {
		cmd = fmt.Sprintf("ip netns exec %s %s", namespace, cmd)
	}
	stdout, _, err := shVerbose(cmd)
	// If err is of type *ExitError then that means it has a non-zero exit
	// code which we are okay with
	if _, ok := err.(*exec.ExitError); !ok && err != nil {
		err = fmt.Errorf("error checking if link %s exists in %s: %s",
			name, namespaceName(namespace), err)
		return false, err
	}
	if string(stdout) == "" {
		return false, nil
	}
	return true, nil
}

// Interprets the empty string as the "root" namespace
func linkDelete(namespace, name string) error {
	if err := ipExec(namespace, "link delete %s", name); err != nil {
		return fmt.Errorf("error deleting link %s in %s: %s",
			name, namespaceName(namespace), err)
	}
	return nil
}

// Lists all veths in the root namespace
func listVeths() ([]string, error) {
	var veths []string

	stdout, _, err := ipExecVerbose("", "link show type veth")
	if err != nil {
		return nil, fmt.Errorf("failed to list veths: %s", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(stdout))
	skipRE := regexp.MustCompile("^\\s+.*")
	vethRE := regexp.MustCompile("^\\d+: (\\w+)@.*")
	for scanner.Scan() {
		line := scanner.Text()
		if skipRE.FindStringIndex(line) != nil {
			// Skip if the line begins with whitespace
			continue
		}
		match := vethRE.FindStringSubmatch(line)
		if match == nil || len(match) != 2 {
			return nil, errors.New("list of veths is not parsing properly")
		}
		veths = append(veths, match[1])
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error while getting veths: %s", err)
	}
	return veths, nil
}

func listIP(namespace, dev string) ([]string, error) {
	var ips []string

	stdout, _, err := ipExecVerbose(namespace, "addr list dev %s", dev)
	if err != nil {
		return nil, fmt.Errorf("failed to list ip addresses in %s: %s",
			namespaceName(namespace), err)
	}

	re, _ := regexp.Compile(`(?:inet|inet6) (\S+)`)
	for _, v := range re.FindAllSubmatch(stdout, -1) {
		ips = append(ips, string(v[1]))
	}

	return ips, nil
}

func addIP(namespace, ip, dev string) error {
	err := ipExec(namespace, "addr add %s dev %s", ip, dev)
	if err != nil {
		return fmt.Errorf("failed to add ip %s to %s in %s: %s",
			ip, dev, namespaceName(namespace), err)
	}
	return nil
}

func delIP(namespace, ip, dev string) error {
	err := ipExec(namespace, "addr del %s dev %s", ip, dev)
	if err != nil {
		return fmt.Errorf("failed to delete ip %s in %s: %s",
			ip, namespaceName(namespace), err)
	}
	return nil
}
