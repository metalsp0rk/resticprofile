//+build windows

package priority

import (
	"fmt"

	"github.com/creativeprojects/resticprofile/clog"
	"golang.org/x/sys/windows"
)

// SetNice sets the unix "nice" value
func SetNice(priority int) error {
	if priority < 0 || priority > 19 {
		return fmt.Errorf("Unexpected priority value %d", priority)
	}
	var class uint32 = windows.NORMAL_PRIORITY_CLASS
	if priority == 19 {
		class = windows.IDLE_PRIORITY_CLASS
	} else if priority == -20 {
		class = windows.HIGH_PRIORITY_CLASS
	} else if priority > 0 {
		class = windows.BELOW_NORMAL_PRIORITY_CLASS
	} else if priority < 0 {
		class = windows.ABOVE_NORMAL_PRIORITY_CLASS
	}
	err := setPriorityClass(class)
	if err != nil {
		return err
	}
	return nil
}

// SetClass sets the priority class of the current process
func SetClass(class int) error {
	switch class {
	case Idle:
		return setPriorityClass(windows.IDLE_PRIORITY_CLASS)
	case Background:
		return setPriorityClass(windows.PROCESS_MODE_BACKGROUND_BEGIN)
	case Low:
		return setPriorityClass(windows.BELOW_NORMAL_PRIORITY_CLASS)
	case Normal:
		return setPriorityClass(windows.NORMAL_PRIORITY_CLASS)
	case High:
		return setPriorityClass(windows.ABOVE_NORMAL_PRIORITY_CLASS)
	case Highest:
		return setPriorityClass(windows.HIGH_PRIORITY_CLASS)
	}
	return fmt.Errorf("Unknown priority class %d", class)
}

func setPriorityClass(class uint32) error {
	handle, err := windows.GetCurrentProcess()
	if err != nil {
		return fmt.Errorf("Error getting current process handle: %v", err)
	}
	clog.Debugf("Setting priority class 0x%x", class)
	err = windows.SetPriorityClass(handle, class)
	if err != nil {
		return fmt.Errorf("Error setting priority class: %v", err)
	}
	return nil
}