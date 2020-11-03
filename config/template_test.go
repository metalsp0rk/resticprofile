package config

import (
	"testing"

	"github.com/creativeprojects/clog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveLock(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	testConfig := `
[profile1]
lock = "/tmp/{{ .Profile.Name }}.lock"
`
	profile, err := getProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	err = ResolveProfileTemplate(profile)
	require.NoError(t, err)
	assert.Equal(t, "/tmp/profile1.lock", profile.Lock)
}

func TestResolveRemainMap(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	testConfig := `
[profile1]
something = "{{ .Profile.Name }}"
`
	profile, err := getProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	err = ResolveProfileTemplate(profile)
	require.NoError(t, err)
	assert.Equal(t, "/tmp/profile1.lock", profile.OtherFlags["something"])
}
