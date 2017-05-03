package google

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProjectID(t *testing.T) {
	_, err := getProjectID("malformed")
	assert.Error(t, err, "")

	_, err = getProjectID(`{"no": "project_id"}`)
	assert.Error(t, err, "")

	id, err := getProjectID(`{"project_id": "myid"}`)
	assert.NoError(t, err)
	assert.Equal(t, "myid", id)
}
