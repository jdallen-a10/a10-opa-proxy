//
//  a10_system.go tests
//

package axapi

import (
	"fmt"
	"testing"
)

func TestGetUptime(t *testing.T) {
	d := setup()
	f, err := d.GetUptime()
	notErr(t, err)
	fmt.Println(f)
}

func TestGetLastConfigSave(t *testing.T) {
	d := setup()
	f, err := d.GetLastConfigSave()
	notErr(t, err)
	fmt.Println("LastConfigSave: " + f)
}

func TestTimezoneCalls(t *testing.T) {
	tz := "America/Toronto"
	d := setup()
	f, err := d.GetTimezone()
	notErr(t, err)
	err = d.SetTimezone(tz)
	notErr(t, err)
	g, err := d.GetTimezone()
	notErr(t, err)
	assert(t, g, tz)
	fmt.Println("Testing Timezone = " + g)
	err = d.SetTimezone(f)
	notErr(t, err)
	g, err = d.GetTimezone()
	notErr(t, err)
	assert(t, g, f)
}

func TestGetProcesses(t *testing.T) {
	d := setup()
	f, err := d.GetProcessInfo()
	notErr(t, err)
	arrContains(t, f, "a10lb")
	arrContains(t, f, "a10mon")
	arrContains(t, f, "a10logd")
}
