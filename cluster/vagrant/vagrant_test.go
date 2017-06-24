package vagrant

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/quilt/quilt/cluster/machine"
)

func TestSetACLs(t *testing.T) {
	clst := Cluster{}
	assert.Nil(t, clst.SetACLs(nil))
}

func TestPreemptibleError(t *testing.T) {
	err := Cluster{}.Boot([]machine.Machine{{Preemptible: true}})
	assert.EqualError(t, err, "vagrant does not support preemptible instances")
}
