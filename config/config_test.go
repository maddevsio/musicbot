package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWeCanGetConfiguration(t *testing.T) {
	cr := New()
	os.Clearenv()
	os.Setenv("HTTP_BIND_ADDR", ":9090")
	cr.Run()
	conf := cr.GetConfig()
	assert.Equal(t, conf.HTTPBindAddr, ":9090")
}
