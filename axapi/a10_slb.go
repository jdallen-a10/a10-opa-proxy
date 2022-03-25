//
//  a10_slb.go  --  SLB related aXAPI API calls
//
//  John D. Allen
//  Sr. Solutions Engineer
//  A10 Networks, Inc.
//
//  Copyright A10 Networks (c) 2020, All Rights Reserved.
//

package axapi

import (
	"strings"

	"github.com/tidwall/gjson"
)

//  GetSLBservers()
//-----------------------------------------------------------------------------
type Server struct {
	Name            string
	Host            string
	Status          string
	Template        string
	ConnectionLimit uint64
	Weight          int
}

// GetSLBservers()
//-----------------------------------------------------------------------------
func (d Device) GetSLBservers() ([]Server, error) {
	var s []Server
	body, err := _restCall(d, "/slb/server", "GET", nil)
	if err != nil {
		return s, err
	}
	if e, msg := d.chkResp(body); e {
		return s, msg
	}

	for _, v := range gjson.GetBytes(body, "server-list").Array() {
		var x Server
		x.Name = gjson.Get(v.String(), "name").Str
		x.Host = gjson.Get(v.String(), "host").Str
		x.Status = gjson.Get(v.String(), "action").Str
		x.Template = gjson.Get(v.String(), "template-server").Str
		x.ConnectionLimit = gjson.Get(v.String(), "conn-limit").Uint()
		x.Weight = int(gjson.Get(v.String(), "weight").Int())
		s = append(s, x)
	}

	return s, nil
}

// GetServceGroups()
//-----------------------------------------------------------------------------
type Member struct {
	Name     string
	Port     int
	State    string
	Priority int
}

type SvcGrp struct {
	Name        string
	Protocol    string
	LBMethod    string
	Healthcheck string
	Members     []Member
}

func (d Device) GetServiceGroups() ([]SvcGrp, error) {
	var sg []SvcGrp
	body, err := _restCall(d, "/slb/service-group-list", "GET", nil)
	if err != nil {
		return sg, err
	}
	if e, msg := d.chkResp(body); e {
		return sg, msg
	}

	for _, v := range gjson.GetBytes(body, "service-group-list").Array() {
		var x SvcGrp
		x.Name = gjson.Get(v.String(), "name").Str
		x.Protocol = gjson.Get(v.String(), "protocol").Str
		x.LBMethod = gjson.Get(v.String(), "lb-method").Str
		x.Healthcheck = gjson.Get(v.String(), "health-check").Str
		for _, z := range gjson.Get(v.String(), "member-list").Array() {
			var m Member
			m.Name = gjson.Get(z.String(), "name").Str
			m.Port = int(gjson.Get(z.String(), "port").Int())
			m.State = gjson.Get(z.String(), "member-state").Str
			m.Priority = int(gjson.Get(z.String(), "member-priority").Int())
			x.Members = append(x.Members, m)
		}
		sg = append(sg, x)
	}

	return sg, nil
}

// GetVSlist()
//-----------------------------------------------------------------------------
type Port struct {
	PortNumber int
	Protocol   string
	ConnLimit  uint64
	Status     string
	AutoSNAT   int
	SvcGrp     string
	Throughput uint64
}

type VS struct {
	Name   string
	IP     string
	Status string
	Ports  []Port
}

func (d Device) GetVSlist() ([]VS, error) {
	var vsl []VS
	body, err := _restCall(d, "/slb/virtual-server-list", "GET", nil)
	if err != nil {
		return vsl, err
	}
	if e, msg := d.chkResp(body); e {
		return vsl, msg
	}

	for _, s := range gjson.GetBytes(body, "virtual-server-list").Array() {
		var vs VS
		vs.Name = gjson.Get(s.String(), "name").Str
		vs.IP = gjson.Get(s.String(), "ip-address").Str
		vs.Status = gjson.Get(s.String(), "enable-disable-action").Str
		for _, v := range gjson.Get(s.String(), "port-list").Array() {
			var p Port
			p.PortNumber = int(gjson.Get(v.String(), "port-number").Int())
			p.Protocol = gjson.Get(v.String(), "protocol").Str
			p.ConnLimit = gjson.Get(v.String(), "conn-limit").Uint()
			p.Status = gjson.Get(v.String(), "action").Str
			p.AutoSNAT = int(gjson.Get(v.String(), "auto").Int())
			p.SvcGrp = gjson.Get(v.String(), "service-group").Str
			vs.Ports = append(vs.Ports, p)
		}
		vsl = append(vsl, vs)
	}

	return vsl, nil
}

// GetVSThroughput()
//-----------------------------------------------------------------------------
func (d Device) GetVSThroughput(vs string) ([]Port, error) {
	// Throughput returned is in bps
	var ports []Port
	url := "/slb/virtual-server/" + vs + "/stats"
	body, err := _restCall(d, url, "GET", nil)
	if err != nil {
		return ports, err
	}
	if e, msg := d.chkResp(body); e {
		return ports, msg
	}

	for _, p := range gjson.GetBytes(body, "virtual-server.port-list").Array() {
		var px Port
		px.PortNumber = int(gjson.Get(p.String(), "port-number").Int())
		px.Protocol = gjson.Get(p.String(), "protocol").Str
		px.Throughput = gjson.Get(p.String(), "stats.throughput-bits-per-sec").Uint()

		ports = append(ports, px)
	}

	return ports, nil
}

// GetServerTemplate()
//-----------------------------------------------------------------------------
func (d Device) GetServerTemplate(tpl string) (string, error) {
	url := "/slb/template/server/" + tpl
	body, err := _restCall(d, url, "GET", nil)
	if err != nil {
		return "", err
	}
	if e, msg := d.chkResp(body); e {
		return "", msg
	}
	return string(body), err
}

// GetVirtualServerTemplate()
//-----------------------------------------------------------------------------
func (d Device) GetVirtualServerTemplate(tpl string) (string, error) {
	url := "/slb/template/virtual-server/" + tpl
	body, err := _restCall(d, url, "GET", nil)
	if err != nil {
		return "", err
	}
	if e, msg := d.chkResp(body); e {
		return "", msg
	}
	return string(body), err
}

// CreateServerTemplate()
//-----------------------------------------------------------------------------
func (d Device) CreateServerTemplate(payload string) error {
	// Payload should have at least the 'name' field, and any attributes you want to set.
	// Example:
	// "server": {
	//     "name": "test",
	//     "conn-limit": 64000000,
	//     "conn-limit-no-logging": 0,
	//     "dns-query-interval": 10,
	//     "dns-fail-interval": 30,
	//     "dynamic-server-prefix": "DRS",
	//     "extended-stats": 0,
	//     "log-selection-failure": 0,
	//     "health-check-disable": 0,
	//     "max-dynamic-server": 255,
	//     "min-ttl-ratio": 2,
	//     "weight": 1,
	//     "spoofing-cache": 0,
	//     "stats-data-action": "stats-data-enable",
	//     "slow-start": 0,
	//     "bw-rate-limit-acct": "all",
	//     "bw-rate-limit": 1000,
	//     "bw-rate-limit-resume": 800,
	//     "bw-rate-limit-duration": 20,
	//     "bw-rate-limit-no-logging": 0
	// }
	url := "/slb/template/server"
	pl := strings.NewReader(payload)
	body, err := _restCall(d, url, "POST", pl)
	if err != nil {
		return err
	}
	if e, msg := d.chkResp(body); e {
		return msg
	}
	return nil
}

// UpdateServerTemplate()
//-----------------------------------------------------------------------------
func (d Device) UpdateServerTemplate(payload string) error {
	// NOTE: The 'name' field MUST be a part of the payload!
	url := "/slb/template/server"
	pl := strings.NewReader(payload)
	body, err := _restCall(d, url, "PUT", pl)
	if err != nil {
		return err
	}
	if e, msg := d.chkResp(body); e {
		return msg
	}
	return nil
}

// CreateVirtualServerTemplate()
//-----------------------------------------------------------------------------
// Payload will at least need the 'name' field, and any KVs you want to set.
// Example:
// "virtual-server": {
// 	  "name": "test2",
// 	  "conn-limit": 200,
// 	  "conn-rate-limit": 200
// }
func (d Device) CreateVirtualServerTemplate(payload string) error {
	url := "/slb/template/virtual-server"
	pl := strings.NewReader(payload)
	body, err := _restCall(d, url, "POST", pl)
	if err != nil {
		return err
	}
	if e, msg := d.chkResp(body); e {
		return msg
	}
	return nil
}

// UpdateVirtualServerTemplate()
//-----------------------------------------------------------------------------
func (d Device) UpdateVirtualServerTemplate(payload string) error {
	// NOTE: The 'name' field MUST be a part of the payload!
	url := "/slb/template/virtual-server"
	pl := strings.NewReader(payload)
	body, err := _restCall(d, url, "PUT", pl)
	if err != nil {
		return err
	}
	if e, msg := d.chkResp(body); e {
		return msg
	}
	return nil
}

// UpdateVirtualServer()
//-----------------------------------------------------------------------------
// This function only adds/updates to a virtual-server. It will overwrite vaules
// if they already exist, or add KV lines to the virtual-server config. It retains
// all other vaules (unlike a PUT would.)
func (d Device) UpdateVirtualServer(vs string, payload string) error {
	url := "/slb/virtual-server/" + vs
	pl := strings.NewReader(payload)
	body, err := _restCall(d, url, "POST", pl)
	if err != nil {
		return err
	}
	if e, msg := d.chkResp(body); e {
		return msg
	}
	return nil
}
