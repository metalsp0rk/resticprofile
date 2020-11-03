package config

import (
	"strconv"
	"testing"
	"time"

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
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Equal(t, "/tmp/profile1.lock", profile.Lock)
}

func TestResolveYear(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	testConfig := `
[profile1]
cache-dir = "{{ .Now.Year }}"
`
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Equal(t, strconv.Itoa(time.Now().Year()), profile.CacheDir)
}

func TestResolveSliceValue(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	testConfig := `
[profile1]
run-before = ["echo {{ .Profile.Name }}"]
`
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Equal(t, "echo profile1", profile.RunBefore[0])
}

func TestResolveStructThenSliceValue(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	testConfig := `
[profile1]
[profile1.backup]
run-before = ["echo {{ .Profile.Name }}"]
`
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Equal(t, "echo profile1", profile.Backup.RunBefore[0])
}

func TestResolveStructThenSliceTwoValues(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	testConfig := `
[profile1]
[profile1.backup]
run-before = ["echo {{ .Profile.Name }}", "ls -al"]
`
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Contains(t, profile.Backup.RunBefore, "echo profile1")
}

func TestYamlResolveStructThenSliceValue(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	testConfig := `---
profile1:
  backup:
    run-before:
      - echo {{ .Profile.Name }}
`
	profile, err := getResolvedProfile("yaml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Equal(t, "echo profile1", profile.Backup.RunBefore[0])
}

func TestYamlResolveStructThenSliceTwoValue(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	testConfig := `---
profile1:
  backup:
    run-before:
      - echo {{ .Profile.Name }}
      - ls -al
`
	profile, err := getResolvedProfile("yaml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Contains(t, profile.Backup.RunBefore, "echo profile1")
}

func TestResolveRemainMap(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	testConfig := `
[profile1]
something = "{{ .Profile.Name }}"
`
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Equal(t, "profile1", profile.OtherFlags["something"])
}

func TestResolveStructThenRemainMap(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	testConfig := `
[profile1]
[profile1.backup]
something = "{{ .Profile.Name }}"
`
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Equal(t, "profile1", profile.Backup.OtherFlags["something"])
}

func TestResolveInheritanceOfProfileName(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	testConfig := `
[profile1]
lock = "/tmp/{{ .Profile.Name }}.lock"
[profile2]
inherit = "profile1"
`
	profile, err := getResolvedProfile("toml", testConfig, "profile2")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Equal(t, "/tmp/profile2.lock", profile.Lock)
}
