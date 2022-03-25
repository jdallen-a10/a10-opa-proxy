//
//  a10_system.go  --  System related aXAPI API calls
//
//  John D. Allen
//  Sr. Solutions Engineer
//  A10 Networks, Inc.
//
//  Copyright A10 Networks (c) 2020, All Rights Reserved.
//

package axapi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

// GetUptime returns the Thunder's uptime as a string
//-----------------------------------------------------------------------------
func (d Device) GetUptime() (string, error) {
	body, err := _restCall(d, "/version/oper", "GET", nil)
	if err != nil {
		return "", err
	}

	if e, msg := d.chkResp(body); e {
		return "", msg
	}

	return gjson.GetBytes(body, "version.oper.up-time").Str, nil
}

// GetPlatform -- What hardware/software is Thunder running on?
//-----------------------------------------------------------------------------
func (d Device) GetPlatform() (string, error) {
	body, err := _restCall(d, "/version/oper", "GET", nil)
	if err != nil {
		return "", err
	}

	if e, msg := d.chkResp(body); e {
		return "", msg
	}

	return gjson.GetBytes(body, "version.oper.hw-platform").Str, nil
}

// BootInfo holds Thunder HD partition info
type BootInfo struct {
	BootFrom  string
	Primary   string
	Secondary string
}

// GetBootInfo - Return struct with info about the two partitions, and which one is booting from.
//-----------------------------------------------------------------------------
func (d Device) GetBootInfo() (BootInfo, error) {
	var b BootInfo
	body, err := _restCall(d, "/bootimage/oper", "GET", nil)
	if err != nil {
		return BootInfo{}, nil
	}

	if e, msg := d.chkResp(body); e {
		return BootInfo{}, msg
	}

	b.BootFrom = gjson.GetBytes(body, "bootimage.hd-default").Str
	b.Primary = gjson.GetBytes(body, "bootimage.hd-pri").Str
	b.Secondary = gjson.GetBytes(body, "bootimage.hd-sec").Str
	return b, nil
}

// GetLastConfigSave - Returns string with time/date of last config save (IE> mem wr)
//-----------------------------------------------------------------------------
func (d Device) GetLastConfigSave() (string, error) {
	body, err := _restCall(d, "/version/oper", "GET", nil)
	if err != nil {
		return "", err
	}

	if e, msg := d.chkResp(body); e {
		return "", msg
	}

	return gjson.GetBytes(body, "version.oper.last-config-saved-time").Str, nil
}

// GetControlCPUs - Returns number of control CPUs
//-----------------------------------------------------------------------------
func (d Device) GetControlCPUs() (int, error) {
	body, err := _restCall(d, "/version/oper", "GET", nil)
	if err != nil {
		return 0, err
	}

	if e, msg := d.chkResp(body); e {
		return 0, msg
	}

	return int(gjson.GetBytes(body, "version.oper.nun-control-cpus").Int()), nil
}

// GetTimezone - What is the Timezone setting on the Thunder device.
//-----------------------------------------------------------------------------
func (d Device) GetTimezone() (string, error) {
	body, err := _restCall(d, "/timezone/oper", "GET", nil)
	if err != nil {
		return "", err
	}

	if e, msg := d.chkResp(body); e {
		return "", msg
	}

	return gjson.GetBytes(body, "timezone.oper.location").Str, nil
}

// SetTimezone - Set the TZ string
//-----------------------------------------------------------------------------
func (d Device) SetTimezone(tz string) error {
	payload := strings.NewReader("{ \"timezone\": {\"timezone-index-cfg\": {\"timezone-index\": \"" + tz + "\" } } }")

	body, err := _restCall(d, "/timezone", "POST", payload)
	if err != nil {
		return err
	}

	if e, msg := d.chkResp(body); e {
		return msg
	}

	return nil
}

// SetHostname - Set the Hostname for the Thunder device
//-----------------------------------------------------------------------------
func (d Device) SetHostname(hn string) error {
	payload := strings.NewReader("{ \"hostname\": {\"value\": \"" + hn + "\" } }")

	body, err := _restCall(d, "/hostname", "PUT", payload)
	if err != nil {
		return err
	}

	if e, msg := d.chkResp(body); e {
		return msg
	}

	return nil
}

// GetProcessInfo - Get the list of running processes
//-----------------------------------------------------------------------------
func (d Device) GetProcessInfo() ([]string, error) {
	var rr []string
	body, err := _restCall(d, "/system-view/show-process/oper", "GET", nil)
	if err != nil {
		return []string{}, err
	}

	if e, msg := d.chkResp(body); e {
		return []string{}, msg
	}

	g := int(gjson.GetBytes(body, "show-process.oper.proc-info.#").Int())
	for i := 0; i < g; i++ {
		t := "show-process.oper.proc-info." + strconv.Itoa(i) + ".proc-data"
		p := gjson.GetBytes(body, t).Str
		r := strings.Split(p, " ")
		if r[2] == "running" {
			rr = append(rr, r[0])
		}
	}
	return rr, nil
}

// CliDeploy - Run a CLI command via the API call
//-----------------------------------------------------------------------------
func (d Device) CliDeploy(cmd string) (string, error) {
	c := strings.NewReader(cmd)
	body, err := _restCall(d, "/clideploy", "POST", c)
	if err != nil {
		return "", err
	}

	if gjson.Valid(string(body)) {
		if e, msg := d.chkResp(body); e {
			return "", msg
		}
		fmt.Println(">>" + string(body))
		panic(string(body))
	} else {
		return string(body), nil
	}
}
