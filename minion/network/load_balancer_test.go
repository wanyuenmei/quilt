package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/quilt/quilt/db"
	"github.com/quilt/quilt/minion/ipdef"
	"github.com/quilt/quilt/minion/ovsdb"
	"github.com/quilt/quilt/minion/ovsdb/mocks"
)

func TestUpdateLoadBalancerIPs(t *testing.T) {
	client := new(mocks.Client)

	// Test error handling.
	client.On("ListLoadBalancers").Return(nil, assert.AnError).Once()
	updateLoadBalancerIPs(client, nil)
	client.AssertNotCalled(t, "CreateLoadBalancer",
		mock.Anything, mock.Anything, mock.Anything)
	client.AssertNotCalled(t, "DeleteLoadBalancer", mock.Anything, mock.Anything)

	// Test joining load balancers.
	client.On("ListLoadBalancers").Return([]ovsdb.LoadBalancer{
		{
			Name: "red",
			VIPs: map[string]string{"10.0.0.2": "10.0.0.3,10.0.0.4"},
		},
		{
			Name: "bad",
		},
	}, nil).Once()
	client.On("DeleteLoadBalancer",
		lSwitch, ovsdb.LoadBalancer{Name: "bad"}).Return(nil)
	client.On("CreateLoadBalancer", lSwitch, "new",
		map[string]string{"10.0.0.10": "10.0.0.11"}).Return(nil)
	updateLoadBalancerIPs(client, []db.Label{
		{
			Label:        "red",
			IP:           "10.0.0.2",
			ContainerIPs: []string{"10.0.0.4", "10.0.0.3"},
		},
		{
			Label:        "new",
			IP:           "10.0.0.10",
			ContainerIPs: []string{"10.0.0.11"},
		},
	})
	client.AssertExpectations(t)
}

func TestUpdateLoadBalancerARP(t *testing.T) {
	client := new(mocks.Client)

	// Test error handling.
	client.On("ListSwitchPort",
		mock.Anything).Return(ovsdb.SwitchPort{}, assert.AnError).Once()
	updateLoadBalancerARP(client, nil)
	client.AssertNotCalled(t, "UpdateSwitchPortAddresses",
		mock.Anything, mock.Anything)

	// Test properly replaces wrong ARP mapping.
	client.On("ListSwitchPort", mock.Anything).Return(ovsdb.SwitchPort{
		Addresses: []string{ipdef.LoadBalancerMac + " 10.0.0.2"},
	}, nil).Once()
	client.On("UpdateSwitchPortAddresses", loadBalancerSwitchPort,
		[]string{ipdef.LoadBalancerMac + " 10.0.0.2 10.0.0.3"}).Return(nil).Once()
	updateLoadBalancerARP(client, []db.Label{
		{IP: "10.0.0.2"}, {IP: "10.0.0.3"},
	})
	client.AssertExpectations(t)

	// Test ignores order IP ordering.
	client.Calls = nil
	client.On("ListSwitchPort", mock.Anything).Return(ovsdb.SwitchPort{
		Addresses: []string{ipdef.LoadBalancerMac + " 10.0.0.2 10.0.0.3"},
	}, nil).Once()
	updateLoadBalancerARP(client, []db.Label{
		{IP: "10.0.0.3"}, {IP: "10.0.0.2"},
	})
	client.AssertNotCalled(t, "UpdateSwitchPortAddresses",
		mock.Anything, mock.Anything)
}
