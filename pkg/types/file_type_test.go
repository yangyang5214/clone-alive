package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsStaticFile(t *testing.T) {
	assert.Equal(t, IsStaticFile("https://202.3.166.101/SAAS/jersey/manager/api/images/5101"), false)
}
