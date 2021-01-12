package crond

import (
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmptyUserCrontab(t *testing.T) {
	crontab := NewCrontab("empty", nil, false)
	buffer := &strings.Builder{}
	err := crontab.Generate(buffer)
	require.NoError(t, err)
	assert.Equal(t, startMarker+endMarker, buffer.String())
}

func TestEmptyRootCrontab(t *testing.T) {
	crontab := NewCrontab("empty", nil, true)
	buffer := &strings.Builder{}
	err := crontab.Generate(buffer)
	require.NoError(t, err)
	assert.Equal(t, "", buffer.String())
}

func TestSimpleUserCrontab(t *testing.T) {
	crontab := NewCrontab("simple", []Entry{NewEntry(calendar.NewEvent(func(event *calendar.Event) {
		event.Minute.MustAddValue(1)
		event.Hour.MustAddValue(1)
	}), "resticprofile backup")}, false)
	buffer := &strings.Builder{}
	err := crontab.Generate(buffer)
	require.NoError(t, err)
	assert.Equal(t, startMarker+"01 01 * * *\tresticprofile backup\n"+endMarker, buffer.String())
}

func TestSimpleRootCrontab(t *testing.T) {
	crontab := NewCrontab("simple", []Entry{NewEntry(calendar.NewEvent(func(event *calendar.Event) {
		event.Minute.MustAddValue(1)
		event.Hour.MustAddValue(1)
	}), "resticprofile backup")}, true)
	buffer := &strings.Builder{}
	err := crontab.Generate(buffer)
	require.NoError(t, err)
	assert.Equal(t, "01 01 * * *\troot\tresticprofile backup\n", buffer.String())
}
