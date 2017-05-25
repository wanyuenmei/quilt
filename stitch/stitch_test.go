package stitch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMissingNode(t *testing.T) {
	lookPath = func(_ string) (string, error) {
		return "", assert.AnError
	}
	_, err := FromFile("unused")
	assert.Error(t, err)
}
