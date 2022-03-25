//
//  a10_network.go  --  System related aXAPI API calls
//
//  John D. Allen
//  Sr. Solutions Engineer
//  A10 Networks, Inc.
//
//  Copyright A10 Networks (c) 2020, All Rights Reserved.
//

package axapi

import (
	"errors"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

// NetInterface holds network interface info. Since it can support multiple interface types, I haven't set it with JSON defaults.
type NetInterface struct {
	IfNum       int
	Name        string
	IPv4Address string
	IPv4Netmask string
	IPv4Gateway string
	IPv6Address string
	IPv6Netmask string
	IPv6Gateway string
	MTU         int
	Status      string
}

// DNS holds all the DNS settings for the Thunder Device
type DNS struct {
	PriIPv4 string
	PriIPv6 string
	SecIPv4 string
	SecIPv6 string
	Suffix  string
}

// GetMgmtIntInfo -- Get info on the Management interface config
//-----------------------------------------------------------------------------
func (d Device) GetMgmtIntInfo() (NetInterface, error) {
	ni := NetInterface{}
	body, err := _restCall(d, "/interface/management", "GET", nil)
	if err != nil {
		return ni, err
	}

	if e, msg := d.chkResp(body); e {
		return ni, msg
	}

	ni.Status = "enable"
	ni.IPv4Address = gjson.GetBytes(body, "management.ip.ipv4-address").Str
	ni.IPv4Netmask = gjson.GetBytes(body, "management.ip.ipv4-netmask").Str
	ni.IPv4Gateway = gjson.GetBytes(body, "management.ip.default-gateway").Str

	return ni, nil
}

// GetIntInfo -- Get info on specified Network Interface
//-----------------------------------------------------------------------------
func (d Device) GetIntInfo(ni NetInterface) (NetInterface, error) {
	url := "/interface/ethernet/" + strconv.Itoa(ni.IfNum)
	body, err := _restCall(d, url, "GET", nil)
	if err != nil {
		return ni, err
	}

	if e, msg := d.chkResp(body); e {
		return ni, msg
	}

	ni.Status = gjson.GetBytes(body, "ethernet.action").Str
	ni.MTU = int(gjson.GetBytes(body, "ethernet.mtu").Int())
	if gjson.GetBytes(body, "ethernet.ip").Exists() {
		ni.IPv4Address = gjson.GetBytes(body, "ethernet.ip.address-list.0.ipv4-address").Str
		ni.IPv4Netmask = gjson.GetBytes(body, "ethernet.ip.address-list.0.ipv4-netmask").Str
	}

	return ni, nil
}

// EnableInt -- Enable the specified Network Interface
//-----------------------------------------------------------------------------
func (d Device) EnableInt(ni NetInterface) (NetInterface, error) {
	url := "/interface/ethernet/" + strconv.Itoa(ni.IfNum)
	payload := strings.NewReader("{ \"ethernet\": { \"ifnum\": " + strconv.Itoa(ni.IfNum) + ", \"action\": \"enable\" } }")
	body, err := _restCall(d, url, "POST", payload)
	if err != nil {
		return ni, err
	}

	if e, msg := d.chkResp(body); e {
		return ni, msg
	}

	ni.Status = gjson.GetBytes(body, "ethernet.action").Str
	return ni, nil
}

// DisableInt -- Enable the specified Network Interface
//-----------------------------------------------------------------------------
func (d Device) DisableInt(ni NetInterface) (NetInterface, error) {
	url := "/interface/ethernet/" + strconv.Itoa(ni.IfNum)
	payload := strings.NewReader("{ \"ethernet\": { \"ifnum\": " + strconv.Itoa(ni.IfNum) + ", \"action\": \"disable\" } }")
	body, err := _restCall(d, url, "POST", payload)
	if err != nil {
		return ni, err
	}

	if e, msg := d.chkResp(body); e {
		return ni, msg
	}

	ni.Status = gjson.GetBytes(body, "ethernet.action").Str
	return ni, nil
}

// SetIntIPv4Address - Set an IPv4 Address on the specified Network Interface
//-----------------------------------------------------------------------------
func (d Device) SetIntIPv4Address(ni NetInterface) (NetInterface, error) {
	if ni.IPv4Address == "" {
		return ni, errors.New("IPv4Address field not set")
	}
	if ni.IPv4Netmask == "" {
		return ni, errors.New("IPv4Netmask field not set")
	}
	url := "/interface/ethernet/" + strconv.Itoa(ni.IfNum)

	payload := strings.NewReader("") // to get around stupid syntax checking
	if ni.Name != "" {
		payload = strings.NewReader("{ \"ethernet\": { \"ifnum\": " + strconv.Itoa(ni.IfNum) + ", \"name\": \"" + ni.Name +
			"\", \"ip\": { \"address-list\": { \"ipv4-address\": \"" + ni.IPv4Address + "\", \"ipv4-netmask\": \"" +
			ni.IPv4Netmask + "\" } } } }")
	} else {
		payload = strings.NewReader("{ \"ethernet\": { \"ifnum\": " + strconv.Itoa(ni.IfNum) +
			", \"ip\": { \"address-list\": { \"ipv4-address\": \"" + ni.IPv4Address + "\", \"ipv4-netmask\": \"" +
			ni.IPv4Netmask + "\" } } } }")
	}

	body, err := _restCall(d, url, "POST", payload)
	if err != nil {
		return ni, err
	}

	if e, msg := d.chkResp(body); e {
		return ni, msg
	}

	ni.Status = gjson.GetBytes(body, "ethernet.action").Str
	ni.MTU = int(gjson.GetBytes(body, "ethernet.mtu").Int())
	ni.IPv4Address = gjson.GetBytes(body, "ethernet.ip.address-list.0.ipv4-address").Str
	ni.IPv4Netmask = gjson.GetBytes(body, "ethernet.ip.address-list.0.ipv4-netmask").Str
	return ni, nil
}

// GetDNSinfo -- Gets all the DNS info from the Thunder device
//-----------------------------------------------------------------------------
func (d Device) GetDNSinfo() (DNS, error) {
	dn := DNS{}
	body, err := _restCall(d, "/ip/dns?detail=true", "GET", nil)
	if err != nil {
		return dn, err
	}

	if e, msg := d.chkResp(body); e {
		return dn, msg
	}

	dn.PriIPv4 = gjson.GetBytes(body, "dns.primary.ip-v4-addr").Str
	dn.PriIPv6 = gjson.GetBytes(body, "dns.primary.ip-v6-addr").Str
	dn.SecIPv4 = gjson.GetBytes(body, "dns.secondary.ip-v4-addr").Str
	dn.SecIPv6 = gjson.GetBytes(body, "dns.secondary.ip-v6-addr").Str
	dn.Suffix = gjson.GetBytes(body, "dns.suffix.domain-name").Str
	return dn, nil
}

// SetPrimaryIPv4DNSserver --  Sets the Primary DNS server
//-----------------------------------------------------------------------------
func (d Device) SetPrimaryIPv4DNSserver(dn DNS) (DNS, error) {
	payload := strings.NewReader("{ \"primary\": { \"ip-v4-addr\": \"" + dn.PriIPv4 + "\" } }")
	body, err := _restCall(d, "/ip/dns/primary", "POST", payload)
	if err != nil {
		return dn, err
	}

	if e, msg := d.chkResp(body); e {
		return dn, msg
	}

	dn.PriIPv4 = gjson.GetBytes(body, "primary.ip-v4-addr").Str
	return dn, nil
}

// SetPrimaryIPv6DNSserver --  Sets the Primary DNS server
//-----------------------------------------------------------------------------
// func (d Device) SetPrimaryIPv6DNSserver(dn DNS) (DNS, error) {
// 	payload := strings.NewReader("{ \"primary\": { \"ip-v6-addr\": \"" + dn.PriIPv6 + "\" } }")
// 	body, err := _restCall(d, "/ip/dns/primary", "POST", payload)
// 	if err != nil {
// 		return dn, err
// 	}

// 	if e, msg := d.chkResp(body); e {
// 		return dn, msg
// 	}

// 	dn.PriIPv6 = gjson.GetBytes(body, "primary.ip-v6-addr").Str
// 	return dn, nil
// }

// SetSecondaryIPv4DNSserver --  Sets the Primary DNS server
//-----------------------------------------------------------------------------
func (d Device) SetSecondaryIPv4DNSserver(dn DNS) (DNS, error) {
	payload := strings.NewReader("{ \"secondary\": { \"ip-v4-addr\": \"" + dn.SecIPv4 + "\" } }")
	body, err := _restCall(d, "/ip/dns/secondary", "POST", payload)
	if err != nil {
		return dn, err
	}

	if e, msg := d.chkResp(body); e {
		return dn, msg
	}

	dn.SecIPv4 = gjson.GetBytes(body, "secondary.ip-v4-addr").Str
	return dn, nil
}

// SetDNSSuffix -- Sets the DNS Search suffix
//-----------------------------------------------------------------------------
func (d Device) SetDNSSuffix(dn DNS) (DNS, error) {
	payload := strings.NewReader("{ \"suffix\": { \"domain-name\": \"" + dn.Suffix + "\" } }")
	body, err := _restCall(d, "/ip/dns/suffix", "POST", payload)
	if err != nil {
		return dn, err
	}

	if e, msg := d.chkResp(body); e {
		return dn, msg
	}

	dn.Suffix = gjson.GetBytes(body, "suffix.domain-name").Str
	return dn, nil
}
