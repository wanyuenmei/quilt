package google

import (
	"errors"
	"testing"

	"github.com/quilt/quilt/cluster/machine"
	"github.com/stretchr/testify/suite"

	compute "google.golang.org/api/compute/v1"
)

type GoogleTestSuite struct {
	suite.Suite

	gce  *mockClient
	clst *Cluster
}

func (s *GoogleTestSuite) SetupTest() {
	s.gce = &mockClient{}
	s.clst = &Cluster{
		gce:  s.gce,
		ns:   "namespace",
		zone: "zone-1",
	}
}

func (s *GoogleTestSuite) TestList() {
	s.gce.On("ListInstances", "zone-1", apiOptions{
		filter: "description eq namespace",
	}).Return(&compute.InstanceList{
		Items: []*compute.Instance{
			{
				MachineType: "machine/split/type-1",
				Name:        "name-1",
				NetworkInterfaces: []*compute.NetworkInterface{
					{
						AccessConfigs: []*compute.AccessConfig{
							{
								NatIP: "x.x.x.x",
							},
						},
						NetworkIP: "y.y.y.y",
					},
				},
			},
		},
	}, nil)

	machines, err := s.clst.List()
	s.NoError(err)
	s.Len(machines, 1)
	s.Equal(machines[0], machine.Machine{
		ID:        "name-1",
		PublicIP:  "x.x.x.x",
		PrivateIP: "y.y.y.y",
		Size:      "type-1",
	})
}

func (s *GoogleTestSuite) TestListFirewalls() {
	s.clst.networkName = "network"
	s.clst.intFW = "intFW"

	s.gce.On("ListFirewalls").Return(&compute.FirewallList{
		Items: []*compute.Firewall{
			{
				Network:    networkURL(s.clst.networkName),
				Name:       "badZone",
				TargetTags: []string{"zone-2"},
			},
			{
				Network:    networkURL(s.clst.networkName),
				Name:       "intFW",
				TargetTags: []string{"zone-1"},
			},
			{
				Network:    networkURL("ignoreMe"),
				Name:       "badNetwork",
				TargetTags: []string{"zone-1"},
			},
			{
				Network:    networkURL(s.clst.networkName),
				Name:       "shouldReturn",
				TargetTags: []string{"zone-1"},
			},
		},
	}, nil).Once()

	fws, err := s.clst.listFirewalls()
	s.NoError(err)
	s.Len(fws, 1)
	s.Equal(fws[0].Name, "shouldReturn")

	s.gce.On("ListFirewalls").Return(nil, errors.New("err")).Once()
	_, err = s.clst.listFirewalls()
	s.EqualError(err, "list firewalls: err")
}

func TestGoogleTestSuite(t *testing.T) {
	suite.Run(t, new(GoogleTestSuite))
}
