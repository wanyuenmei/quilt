package network

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vishvananda/netlink"

	"github.com/quilt/quilt/minion/ipdef"
)

func TestMakeIPRule(t *testing.T) {
	inp := "-A INPUT -p tcp -i eth0 -m multiport --dports 465,110,995 -j ACCEPT"
	rule, _ := makeIPRule(inp)
	expCmd := "-A"
	expChain := "INPUT"
	expOpts := "-p tcp -i eth0 -m multiport --dports 465,110,995 -j ACCEPT"

	if rule.cmd != expCmd {
		t.Errorf("Bad ipRule command.\nExpected:\n%s\n\nGot:\n%s\n",
			expCmd, rule.cmd)
	}

	if rule.chain != expChain {
		t.Errorf("Bad ipRule chain.\nExpected:\n%s\n\nGot:\n%s\n",
			expChain, rule.chain)
	}

	if rule.opts != expOpts {
		t.Errorf("Bad ipRule options.\nExpected:\n%s\n\nGot:\n%s\n",
			expOpts, rule.opts)
	}

	inp = "-A POSTROUTING -s 10.0.3.0/24 ! -d 10.0.3.0/24 -j MASQUERADE"
	rule, _ = makeIPRule(inp)
	expCmd = "-A"
	expChain = "POSTROUTING"
	expOpts = "-s 10.0.3.0/24 ! -d 10.0.3.0/24 -j MASQUERADE"

	if rule.cmd != expCmd {
		t.Errorf("Bad ipRule command.\nExpected:\n%s\n\nGot:\n%s\n",
			expCmd, rule.cmd)
	}

	if rule.chain != expChain {
		t.Errorf("Bad ipRule chain.\nExpected:\n%s\n\nGot:\n%s\n",
			expChain, rule.chain)
	}

	if rule.opts != expOpts {
		t.Errorf("Bad ipRule options.\nExpected:\n%s\n\nGot:\n%s\n",
			expOpts, rule.opts)
	}

	inp = "-A PREROUTING -i eth0 -p tcp --dport 80 -j DNAT " +
		"--to-destination 10.31.0.23:80"
	rule, _ = makeIPRule(inp)
	expCmd = "-A"
	expChain = "PREROUTING"
	expOpts = "-i eth0 -p tcp --dport 80 -j DNAT --to-destination 10.31.0.23:80"

	if rule.cmd != expCmd {
		t.Errorf("Bad ipRule command.\nExpected:\n%s\n\nGot:\n%s\n",
			expCmd, rule.cmd)
	}

	if rule.chain != expChain {
		t.Errorf("Bad ipRule chain.\nExpected:\n%s\n\nGot:\n%s\n",
			expChain, rule.chain)
	}

	if rule.opts != expOpts {
		t.Errorf("Bad ipRule options.\nExpected:\n%s\n\nGot:\n%s\n",
			expOpts, rule.opts)
	}
}

func TestGenerateCurrentNatRules(t *testing.T) {
	oldShVerbose := shVerbose
	defer func() { shVerbose = oldShVerbose }()
	shVerbose = func(format string, args ...interface{}) (
		stdout, stderr []byte, err error) {
		return []byte(rules()), nil, nil
	}

	actual, _ := generateCurrentNatRules()
	exp := ipRuleSlice{
		{
			cmd:   "-P",
			chain: "POSTROUTING",
			opts:  "ACCEPT",
		},
		{
			cmd:   "-N",
			chain: "DOCKER",
		},
		{
			cmd:   "-A",
			chain: "POSTROUTING",
			opts:  "-s 11.0.0.0/8,10.0.0.0/8 -o eth0 -j MASQUERADE",
		},
		{
			cmd:   "-A",
			chain: "POSTROUTING",
			opts:  "-s 10.0.3.0/24 ! -d 10.0.3.0/24 -j MASQUERADE",
		},
	}

	if !(reflect.DeepEqual(actual, exp)) {
		t.Errorf("Generated wrong routes.\nExpected:\n%+v\n\nGot:\n%+v\n",
			exp, actual)
	}
}

func TestGetPublicInterface(t *testing.T) {
	routeList = func(link netlink.Link, family int) ([]netlink.Route, error) {
		return nil, errors.New("not implemented")
	}
	linkByIndex = func(index int) (netlink.Link, error) {
		if index == 5 {
			link := netlink.GenericLink{}
			link.LinkAttrs.Name = "link name"
			return &link, nil
		}
		return nil, errors.New("unknown")
	}

	res, err := getPublicInterface()
	assert.Empty(t, res)
	assert.EqualError(t, err, "route list: not implemented")

	var routes []netlink.Route
	routeList = func(link netlink.Link, family int) ([]netlink.Route, error) {
		return routes, nil
	}
	res, err = getPublicInterface()
	assert.Empty(t, res)
	assert.EqualError(t, err, "missing default route")

	routes = []netlink.Route{{Dst: &ipdef.QuiltSubnet}}
	res, err = getPublicInterface()
	assert.Empty(t, res)
	assert.EqualError(t, err, "missing default route")

	routes = []netlink.Route{{LinkIndex: 2}}
	res, err = getPublicInterface()
	assert.Empty(t, res)
	assert.EqualError(t, err, "default route missing interface: unknown")

	routes = []netlink.Route{{LinkIndex: 5}}
	res, err = getPublicInterface()
	assert.Equal(t, "link name", res)
	assert.NoError(t, err)
}

func rules() string {
	return `-P POSTROUTING ACCEPT
-N DOCKER
-A POSTROUTING -s 11.0.0.0/8,10.0.0.0/8 -o eth0 -j MASQUERADE
-A POSTROUTING -s 10.0.3.0/24 ! -d 10.0.3.0/24 -j MASQUERADE`
}
