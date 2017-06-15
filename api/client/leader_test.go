package client

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/quilt/quilt/api"
	"github.com/quilt/quilt/api/client/mocks"
	"github.com/quilt/quilt/db"
)

func TestLeader(t *testing.T) {
	passedClient := &mocks.Client{}
	newClient = func(host string) (Client, error) {
		switch host {
		// One machine doesn't know the LeaderIP
		case api.RemoteAddress("8.8.8.8"):
			return &mocks.Client{
				EtcdReturn: []db.Etcd{
					{
						LeaderIP: "",
					},
				},
			}, nil
		// The other machine knows the LeaderIP
		case api.RemoteAddress("9.9.9.9"):
			return &mocks.Client{
				EtcdReturn: []db.Etcd{
					{
						LeaderIP: "leader-priv",
					},
				},
			}, nil
		case api.RemoteAddress("leader"):
			return passedClient, nil
		default:
			t.Fatalf("Unexpected call to getClient with host %s",
				host)
		}
		panic("unreached")
	}

	res, err := Leader([]db.Machine{
		{
			PublicIP: "8.8.8.8",
		},
		{
			PublicIP: "9.9.9.9",
		},
		{
			PublicIP:  "leader",
			PrivateIP: "leader-priv",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, passedClient, res)
}

func TestNoLeader(t *testing.T) {
	newClient = func(host string) (Client, error) {
		// No client knows the leader IP.
		return &mocks.Client{
			EtcdReturn: []db.Etcd{
				{
					LeaderIP: "",
				},
			},
		}, nil
	}

	_, err := Leader([]db.Machine{
		{
			PublicIP: "8.8.8.8",
		},
		{
			PublicIP: "9.9.9.9",
		},
	})
	assert.EqualError(t, err, "no leader found")
}
