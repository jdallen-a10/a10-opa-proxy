//
//  axapi.go tests
//

package axapi

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

var Hostname string
var Username string
var Password string
var Address string
var Token string
var Version string
var Hardware string
var BootFrom string
var SerialNumber string

func setup() Device {
	var d Device
	d.Address = Address
	d.Username = Username
	d.Password = Password
	d.Token = Token
	d.Hostname = Hostname
	d.Version = Version
	d.Hardware = Hardware
	d.BootFrom = BootFrom
	d.SerialNumber = SerialNumber
	return d
}

// Testing Helper Functions
//---------------------------------------------------------------------------

// notErr() - Check to see if the returned err is NOT nil
func notErr(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

// isErr()  -  Should return an error condition
func isErr(t *testing.T, err error, msg string) {
	if err == nil {
		t.Error(msg)
	}
}

// assert() if the two values are equal, if not error out
func assert(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		return
	}
	t.Errorf("Received \"%v\" (type %v), expected \"%v\" (type %v)", a, reflect.TypeOf(a), b, reflect.TypeOf(b))
}

// assertNot() that the two values are NOT equal
func assertNot(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		return
	}
	t.Errorf("Received \"%v\" (type %v), but NOT expecting to", a, reflect.TypeOf(a))
}

// contains() - does the string contain somewhere the value?
func contains(t *testing.T, src string, v string) {
	if strings.Contains(src, v) {
		return
	}
	t.Errorf("String \"%s\" does not Contain \"%s\"", src, v)
}

// arrContains() - does the []string contain the value?
func arrContains(t *testing.T, src []string, v string) {
	for i := range src {
		if src[i] == v {
			return
		}
	}
	t.Errorf("Slice \"%v\" does not Contain \"%s\"", src, v)
}

//---------------------------------------------------------------------------
func TestMain(m *testing.M) {
	// --- Log into our test Thunder device ---
	var d Device
	d.Address = os.Getenv("A10IP")
	d.Username = "admin"
	d.Password = "a10"
	d, err := d.Login()
	if err != nil {
		panic(err)
	}
	if d.Token == "" {
		panic("A10 Auth Token Missing")
	}
	// Setup Device vars
	Username = d.Username
	Password = d.Password
	Address = d.Address
	Token = d.Token
	Hostname = ""

	// --- Run the Tests ---
	flag.Parse()
	ex := m.Run()

	// --- Log off the Thunder device ---
	d, err = d.Logoff()
	if err != nil {
		panic(err)
	}

	// --- End the Tests ---
	os.Exit(ex)
}

func TestHostname(t *testing.T) {
	d := setup()
	d, err := d.GetHostname()
	notErr(t, err)
	Hostname := d.Hostname

	err = d.SetHostname("Testing1")
	notErr(t, err)

	d, err = d.GetHostname()
	notErr(t, err)
	assert(t, d.Hostname, "Testing1")

	err = d.SetHostname(Hostname)
	notErr(t, err)

	d, err = d.GetHostname()
	notErr(t, err)
	assert(t, d.Hostname, Hostname)
}

func TestGetVersion(t *testing.T) {
	d := setup()
	d, err := d.GetVersion()
	notErr(t, err)
}

func TestGetVirtType(t *testing.T) {
	d := setup()
	f, err := d.GetVirtType()
	notErr(t, err)
	fmt.Println("VirtType: " + f)
}
