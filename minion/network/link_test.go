package network

import (
	"errors"
	"reflect"
	"testing"

	"github.com/NetSys/quilt/minion/ipdef"
	"github.com/stretchr/testify/assert"
)

func TestListIP(t *testing.T) {
	oldIPExecVerbose := ipExecVerbose
	defer func() { ipExecVerbose = oldIPExecVerbose }()
	ipExecVerbose = func(namespace, format string, args ...interface{}) (
		stdout, stderr []byte, err error) {
		return []byte(ips()), nil, nil
	}
	actual, _ := listIP("", innerVeth)
	exp := []string{"10.0.2.15/24", "fe80::a00:27ff:fe9b:594e/64"}
	if !reflect.DeepEqual(actual, exp) {
		t.Errorf("Generated wrong IPs.\nExpected:\n%s\n\nGot:\n%s\n",
			exp, actual)
	}
}

type execCount struct {
	c int
}

func (ec *execCount) exec(a, b string, c ...interface{}) ([]byte, []byte, error) {
	ec.c++
	return nil, nil, nil
}

func (ec *execCount) execErr(a, b string, c ...interface{}) ([]byte, []byte, error) {
	ec.c++
	return nil, nil, errors.New("err")
}

func (ec *execCount) sh(a string, c ...interface{}) ([]byte, []byte, error) {
	ec.c++
	return nil, nil, nil
}

func (ec *execCount) shErr(a string, c ...interface{}) ([]byte, []byte, error) {
	ec.c++
	return nil, nil, errors.New("err")
}

func TestDelVeth(t *testing.T) {
	ec := &execCount{}
	oldIPExecVerbose := ipExecVerbose
	defer func() { ipExecVerbose = oldIPExecVerbose }()
	ipExecVerbose = ec.exec

	err := DelVeth("0000000000000")
	assert.Nil(t, err)
	assert.Equal(t, 1, ec.c)
}

func TestAddVeth(t *testing.T) {
	ec := &execCount{}
	oldIPExecVerbose := ipExecVerbose
	defer func() { ipExecVerbose = oldIPExecVerbose }()
	ipExecVerbose = ec.exec

	peer, err := AddVeth("0000000000000")
	assert.Nil(t, err)
	assert.Equal(t, 3, ec.c)

	expPeer := ipdef.IFName("tmp_" + "0000000000000")
	assert.Equal(t, expPeer, peer)

	ipExecVerbose = ec.execErr
	peer, err = AddVeth("1111111111111")
	assert.NotNil(t, err)
	assert.Equal(t, 4, ec.c)
	assert.Equal(t, "", peer)
}

func TestLinkExists(t *testing.T) {
	ec := &execCount{}
	oldSHVerbose := shVerbose
	defer func() { shVerbose = oldSHVerbose }()
	shVerbose = ec.sh

	ok, err := LinkExists("", "0000000000000")
	assert.Nil(t, err)
	assert.False(t, ok)
	assert.Equal(t, 1, ec.c)

	shVerbose = ec.shErr
	ok, err = LinkExists("", "0000000000000")
	assert.NotNil(t, err)
	assert.False(t, ok)
	assert.Equal(t, 2, ec.c)
}

func ips() string {
	return `2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000
    link/ether 08:00:27:9b:59:4e brd ff:ff:ff:ff:ff:ff
    inet 10.0.2.15/24 brd 10.0.2.255 scope global eth0
		valid_lft forever preferred_lft forever
    inet6 fe80::a00:27ff:fe9b:594e/64 scope link
		valid_lft forever preferred_lft forever
    6: eth1: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN group default
		link/ether 0e:9f:0c:21:65:4a brd ff:ff:ff:ff:ff:ff`
}
