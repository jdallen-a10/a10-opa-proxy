//
//  a10_network.go tests
//

package axapi

import (
	"fmt"
	"testing"
)

func TestGetMgmtIntInfo(t *testing.T) {
	d := setup()
	n, err := d.GetMgmtIntInfo()
	notErr(t, err)
	assert(t, n.Status, "enable")
	assert(t, n.IPv4Address, d.Address)
}

func TestGetIntInfo(t *testing.T) {
	d := setup()
	ni := NetInterface{}
	ni.IfNum = 1
	ni, err := d.GetIntInfo(ni)
	notErr(t, err)
	assert(t, ni.MTU, 1500)
	assert(t, ni.Status, "disable")
	assertNot(t, ni.IPv4Address, "")
	fmt.Println(ni)
	fmt.Println("IP Address = " + ni.IPv4Address)
}

func TestEnableInt(t *testing.T) {
	d := setup()
	ni := NetInterface{}
	ni.IfNum = 1
	ni, err := d.GetIntInfo(ni)
	notErr(t, err)
	assert(t, ni.MTU, 1500)
	ni, err = d.EnableInt(ni)
	notErr(t, err)
	assert(t, ni.Status, "enable")
}

func TestDisableInt(t *testing.T) {
	d := setup()
	ni := NetInterface{}
	ni.IfNum = 1
	ni, err := d.GetIntInfo(ni)
	notErr(t, err)
	assert(t, ni.MTU, 1500)
	ni, err = d.DisableInt(ni)
	notErr(t, err)
	assert(t, ni.Status, "disable")
}

func TestSetIntIPv4Address(t *testing.T) {
	d := setup()
	ni := NetInterface{}
	ni.IfNum = 1
	ni, err := d.SetIntIPv4Address(ni) // this SHOULD err
	isErr(t, err, "Uncaught test for blank IPv4Address or IPv4Netmask field")
	ni.Name = "test-data-in"
	ni.IPv4Address = "10.1.1.44"
	ni.IPv4Netmask = "255.255.255.0"
	ni, err = d.SetIntIPv4Address(ni)
	notErr(t, err)
	assert(t, ni.IPv4Address, "10.1.1.44")
	assert(t, ni.IPv4Netmask, "255.255.255.0")
}

func TestGetDNSinfo(t *testing.T) {
	d := setup()
	dn, err := d.GetDNSinfo()
	notErr(t, err)
	assert(t, dn.PriIPv4, "44.147.45.53")
	assert(t, dn.SecIPv4, "44.147.45.28")
	assert(t, dn.Suffix, "home.gan")
}

func TestSetPrimaryDNSserver(t *testing.T) {
	d := setup()
	dn := DNS{}
	dn.PriIPv4 = "44.147.45.53"
	dn, err := d.SetPrimaryIPv4DNSserver(dn)
	notErr(t, err)
	assert(t, dn.PriIPv4, "44.147.45.53")
}

// func TestSetPrimaryIPv6DNSserver(t *testing.T) {
// 	d := setup()
// 	dn := DNS{}
// 	dn.PriIPv6 = "2605:6000:8b04:6900:2ac:e0ff:fea6:c957"
// 	dn, err := d.SetPrimaryIPv6DNSserver(dn)
// 	notErr(t, err)
// 	assert(t, dn.PriIPv6, "2605:6000:8b04:6900:2ac:e0ff:fea6:c957")
// }

func TestSetSecondaryDNSserver(t *testing.T) {
	d := setup()
	dn := DNS{}
	dn.SecIPv4 = "44.147.45.28"
	dn, err := d.SetSecondaryIPv4DNSserver(dn)
	notErr(t, err)
	assert(t, dn.SecIPv4, "44.147.45.28")
}

func TestSetDNSSuffix(t *testing.T) {
	d := setup()
	dn := DNS{}
	dn.Suffix = "home.gan"
	dn, err := d.SetDNSSuffix(dn)
	notErr(t, err)
	assert(t, dn.Suffix, "home.gan")
}
