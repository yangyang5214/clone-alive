package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsStaticFile(t *testing.T) {
	assert.Equal(t, IsStaticFile("https://202.3.166.101/SAAS/jersey/manager/api/images/5101"), false)
}
