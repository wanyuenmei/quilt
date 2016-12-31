package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	clientMock "github.com/NetSys/quilt/api/client/mocks"
	"github.com/NetSys/quilt/db"
	"github.com/NetSys/quilt/quiltctl/testutils"
	"github.com/NetSys/quilt/stitch"
)

func TestStopNamespaceDefault(t *testing.T) {
	t.Parallel()

	mockGetter := new(testutils.Getter)
	c := &clientMock.Client{}
	mockGetter.On("Client", mock.Anything).Return(c, nil)

	c.ClusterReturn = []db.Cluster{
		{
			Spec: `{"namespace": "testSpace"}`,
		},
	}

	stopCmd := NewStopCommand()
	stopCmd.clientGetter = mockGetter
	stopCmd.Run()
	assertDeployed(t, stitch.Stitch{Namespace: "testSpace"}, c.DeployArg)

	c.ClusterReturn = nil
	assert.Equal(t, 1, stopCmd.Run(),
		"can't retrieve namespace if no cluster is deployed")
}

func TestStopNamespace(t *testing.T) {
	t.Parallel()

	mockGetter := new(testutils.Getter)
	c := &clientMock.Client{}
	mockGetter.On("Client", mock.Anything).Return(c, nil)

	stopCmd := NewStopCommand()
	stopCmd.clientGetter = mockGetter
	stopCmd.namespace = "namespace"
	stopCmd.Run()

	assertDeployed(t, stitch.Stitch{Namespace: "namespace"}, c.DeployArg)
}

func assertDeployed(t *testing.T, exp stitch.Stitch, deployed string) {
	actual, err := stitch.FromJSON(deployed)
	assert.NoError(t, err)
	assert.Equal(t, exp, actual, "incorrect stop spec deployed")
}

func TestStopFlags(t *testing.T) {
	t.Parallel()

	expNamespace := "namespace"
	checkStopParsing(t, []string{"-namespace", expNamespace}, expNamespace, nil)
	checkStopParsing(t, []string{expNamespace}, expNamespace, nil)
	checkStopParsing(t, []string{}, "", nil)
}

func checkStopParsing(t *testing.T, args []string, expNamespace string, expErr error) {
	stopCmd := NewStopCommand()
	err := parseHelper(stopCmd, args)

	assert.Equal(t, expErr, err)
	assert.Equal(t, expNamespace, stopCmd.namespace)
}
