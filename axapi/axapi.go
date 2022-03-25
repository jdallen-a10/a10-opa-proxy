//
//
//  Makes use of https://github.com/tidwall/gjson
//
//  John D. Allen
//  Sr. Solutions Engineer
//  A10 Networks, Inc.
//
//  Copyright (c) 2020, All Rights Reserved.
//

package axapi

import (
	"crypto/tls"
	//"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
)

// Device holds A10 Thunder node info
//-----------------------------------------------------------------------------
type Device struct {
	Hostname     string
	Username     string
	Password     string
	Address      string
	Token        string
	Version      string
	Hardware     string
	BootFrom     string
	SerialNumber string
}

// _restCall is the basic API callout function
//-----------------------------------------------------------------------------
func _restCall(d Device, url string, method string, payload *strings.Reader) ([]byte, error) {
	var body []byte
	if d.Token == "" && url != "/auth" {
		return []byte{}, errors.New("No A10 Auth Token! You must Login() before calling other API calls")
	}

	u := "https://" + d.Address + "/axapi/v3" + url
	if method == "" {
		method = "GET"
	}
	if payload == nil {
		payload = strings.NewReader("")
	}

	// Skip insecure SSL verify returns -- lots of Thunders don't have this set.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	// set the HTTPS request
	req, err := http.NewRequest(method, u, payload)
	if err != nil {
		return []byte{}, err
	}

	if url == "/clideploy" {
		req.Header.Add("Content-Type", "text/plain")
	} else {
		req.Header.Add("Content-Type", "application/json")
	}
	if d.Token != "" {
		req.Header.Add("Authorization", d.Token)
	}

	res, err := client.Do(req)

	if err != nil {
		return []byte{}, err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	//fmt.Println(res)
	if res.StatusCode > 299 { // Check for API Errors on Call
		return []byte{}, errors.New(res.Status)
	}

	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}

// chkResp -- Check the response from _restCall() for "fail"  response status
//-----------------------------------------------------------------------------
func (d Device) chkResp(b []byte) (bool, error) {
	if gjson.GetBytes(b, "response.status").Exists() {
		if gjson.GetBytes(b, "response.status").Str == "fail" {
			//fmt.Println(string(body))
			return true, errors.New(gjson.GetBytes(b, "response.err.msg").Str)
		}
	}
	return false, nil
}

// Login to the A10 Thunder device
//-----------------------------------------------------------------------------
func (d Device) Login() (Device, error) {
	if d.Username == "" {
		d.Username = "admin"
	}
	if d.Password == "" && d.Username == "admin" {
		d.Password = "a10"
	}

	payload := strings.NewReader("{\n\"credentials\": {\n\"username\": \"" + d.Username + "\",\n\"password\": \"" + d.Password + "\"\n}\n}")

	body, err := _restCall(d, "/auth", "POST", payload)
	if err != nil {
		return d, err
	}

	d.Token = "A10 " + gjson.GetBytes(body, "authresponse.signature").Str
	return d, nil
}

// GetHostname gets the hostname that the Thunder Device has currently assigned.
//-----------------------------------------------------------------------------
func (d Device) GetHostname() (Device, error) {
	body, err := _restCall(d, "/hostname", "GET", nil)
	if err != nil {
		return d, err
	}

	d.Hostname = gjson.GetBytes(body, "hostname.value").Str
	return d, nil
}

// GetVersion retrieves ACOS version info
//-----------------------------------------------------------------------------
func (d Device) GetVersion() (Device, error) {
	body, err := _restCall(d, "/version/oper", "GET", nil)
	if err != nil {
		return d, err
	}

	d.Version = gjson.GetBytes(body, "version.oper.sw-version").Str
	d.Hardware = gjson.GetBytes(body, "version.oper.hw-platform").Str
	d.BootFrom = gjson.GetBytes(body, "version.oper.boot-from").Str
	d.SerialNumber = gjson.GetBytes(body, "version.oper.serial-number").Str
	return d, nil
}

// GetVirtType will return a string of the virtualization type, if available
//-----------------------------------------------------------------------------
func (d Device) GetVirtType() (string, error) {
	body, err := _restCall(d, "/version/oper", "GET", nil)
	if err != nil {
		return "", err
	}

	return gjson.GetBytes(body, "version.oper.virtualization-type").Str, nil
}

// Logoff - terminates the current API session
//-----------------------------------------------------------------------------
func (d Device) Logoff() (Device, error) {
	body, err := _restCall(d, "/logoff", "GET", nil)
	if err != nil {
		return d, err
	}
	if gjson.GetBytes(body, "response.status").Str != "OK" {
		return d, errors.New("Invalid Logoff: " + gjson.GetBytes(body, "response.status").Str)
	}

	d.Token = ""
	return d, nil
}
