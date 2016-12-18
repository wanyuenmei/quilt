package network

import (
	"fmt"
	"regexp"
)

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
