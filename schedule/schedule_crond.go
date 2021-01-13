//+build !darwin,!windows

package schedule

import (
	"os"
	"strings"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/crond"
)

const (
	crontabBin = "crontab"
)

// createCrondJob is creating the crontab
func (j *Job) createCrondJob(schedules []*calendar.Event) error {
	entries := make([]crond.Entry, len(schedules))
	for i, event := range schedules {
		entries[i] = crond.NewEntry(event, j.config.Title(), j.config.SubTitle(), j.config.Command()+" "+strings.Join(j.config.Arguments(), " "))
	}
	crontab := crond.NewCrontab(j.config.Command(), entries)
	crontab.Generate(os.Stdout)
	return nil
}

// displayCrondStatus has nothing to display (crond doesn't provide running information)
func (j *Job) displayCrondStatus(command string) error {
	return nil
}
