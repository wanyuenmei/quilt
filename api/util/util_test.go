package util

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/NetSys/quilt/api/client/mocks"
	"github.com/NetSys/quilt/db"
)

func TestGetContainer(t *testing.T) {
	t.Parallel()

	passedContainer := db.Container{
		StitchID: "1",
	}
	c := &mocks.Client{
		ContainerReturn: []db.Container{
			passedContainer,
			{StitchID: "2"},
			{StitchID: "3"},
		},
	}
	res, err := GetContainer(c, passedContainer.StitchID)
	assert.Nil(t, err)
	assert.Equal(t, passedContainer, res)
}

func TestGetContainerErr(t *testing.T) {
	t.Parallel()

	c := &mocks.Client{
		ContainerReturn: []db.Container{
			{StitchID: "2"},
			{StitchID: "3"},
		},
	}
	_, err := GetContainer(c, "1")
	assert.EqualError(t, err, `no container with stitchID "1"`)
}
