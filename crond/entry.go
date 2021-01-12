package crond

import (
	"fmt"
	"io"
	"strings"

	"github.com/creativeprojects/resticprofile/calendar"
)

type Entry struct {
	event   *calendar.Event
	root    bool
	command string
}

func NewEntry(event *calendar.Event, root bool, command string) *Entry {
	return &Entry{
		event:   event,
		root:    root,
		command: command,
	}
}

// Generate writes a cron line in the StringWriter (end of line included)
func (e *Entry) Generate(w io.StringWriter) error {
	minute, hour, dayOfMonth, month, dayOfWeek := "*", "*", "*", "*", "*"
	user := ""
	if e.root {
		user = "root\t"
	}
	if e.event.Minute.HasValue() {
		minute = formatRange(e.event.Minute.GetRanges(), twoDecimals)
	}
	if e.event.Hour.HasValue() {
		hour = formatRange(e.event.Hour.GetRanges(), twoDecimals)
	}
	if e.event.Day.HasValue() {
		dayOfMonth = formatRange(e.event.Day.GetRanges(), twoDecimals)
	}
	if e.event.Month.HasValue() {
		month = formatRange(e.event.Month.GetRanges(), twoDecimals)
	}
	if e.event.WeekDay.HasValue() {
		// don't make ranges for days of the week as it can fail with high sunday (7)
		dayOfWeek = formatList(e.event.WeekDay.GetRangeValues(), formatWeekDay)
	}
	_, err := w.WriteString(fmt.Sprintf("%s %s %s %s %s\t%s%s\n", minute, hour, dayOfMonth, month, dayOfWeek, user, e.command))
	return err
}

func formatWeekDay(weekDay int) string {
	if weekDay >= 7 {
		weekDay -= 7
	}
	return fmt.Sprintf("%d", weekDay)
}

func twoDecimals(value int) string {
	return fmt.Sprintf("%02d", value)
}

func formatList(values []int, formatter func(int) string) string {
	output := make([]string, len(values))
	for i, value := range values {
		output[i] = formatter(value)
	}
	return strings.Join(output, ",")
}

func formatRange(values []calendar.Range, formatter func(int) string) string {
	output := make([]string, len(values))
	for i, value := range values {
		if value.End-value.Start > 1 || value.Start-value.End > 1 {
			// proper range
			output[i] = formatter(value.Start) + "-" + formatter(value.End)
		} else if value.End != value.Start {
			// contiguous values
			output[i] = formatter(value.Start) + "," + formatter(value.End)
		} else {
			// single value
			output[i] = formatter(value.Start)
		}
	}
	return strings.Join(output, ",")
}
