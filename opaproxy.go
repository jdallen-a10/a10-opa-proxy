package main

//
//  opaproxy.go  --  A Proof-of-Concept Thunder Cloud Agent (TCA) to retrieve Policy from an
//  Open Policy Agent (OPA) [https://www.openpolicyagent.org/] and implement that policy on a
//  defined Thunder node.  For this particular POC, we will be setting Connection Rate Limiting
//  Policy to limit how many connections per second an SLB will allow. We also have placed
//  some hooks in the code to also support a Bandwidth Control Policy for members servers in a
//  Service-Group -- but did not finish the code to implement the policy on the Thunder node
//  in this version....Ran out of alloted time ;)
//
//---------------------------------------------------------------------------------
//  John D. Allen
//  Global Solutions Architect -- Cloud, IoT, & Automation
//  A10 Networks
//  January, 2022
//---------------------------------------------------------------------------------

import (
	"a10/axapi"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

//---------------------------------------------------------------------------------
// Command Line Setting Global Variables
var DEBUG int
var OPA_PORT int
var OPA_IP string
var THND_PORT int
var THND_IP string
var THND_ID string
var CFG_FILE string

//---------------------------------------------------------------------------------
// Configuration struct
// NOTE: The main set of config items are Unmarshalled as YAML, but the 'Virts" item
// is brought in as a JSON array.  Using the 'gopkg.in/yaml.v2' package, it seems to
// pass the JSON just fine into the []Virtual structure.
type Virtual struct {
	Name   string `json:"name"`
	Policy string `json:"policy"`
}

type Configuration struct {
	Debug        int           `yaml:"debug"`
	OPA_IP       string        `yaml:"OPA_IP"`
	OPA_PORT     int           `yaml:"OPA_PORT"`
	THND_IP      string        `yaml:"THND_IP"`
	THND_PORT    int           `yaml:"THND_PORT"`
	THND_USER    string        `yaml:"THND_USER"`
	THND_PASSWD  string        `yaml:"THND_PASSWD"`
	THND_ID      string        `yaml:"THND_ID"`
	Virts        []Virtual     `yaml:"vs"`
	CHK_INTERVAL time.Duration `yaml:"CHECK_INTERVAL"`
}

//---------------------------------------------------------------------------------
// getConfig() - Grab configuration variables from the config YAML file
func getConfig(fn string) (Configuration, error) {
	var c Configuration

	yamlFile, err := ioutil.ReadFile(fn)
	if err != nil {
		return Configuration{}, err
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		return Configuration{}, err
	}

	return c, nil
}

//---------------------------------------------------------------------------------
//  callOPA()  --  Call the OPA Server API
// NOTE:  This is only using standard HTTP, NOT HTTPS!  If you need a more secure
// connection to pass your network policies around, you should implement this as
// an HTTPS connection. Here's one example: https://github.com/jcbsmpsn/golang-https-example
// And Yes, I know there is a GO specific OPA module, but I didn't use it because it
// seems to clash with other modules I was using at the start, so I went with the
// RESTful API instead...its also more flexible this way IHMO. -- John
func callOPA(url string, method string, payload string) (string, error) {
	cc := &http.Client{}
	pp := strings.NewReader(payload)
	req, err := http.NewRequest(method, url, pp)
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	rsp, err := cc.Do(req)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(rsp.Body)
	out := buf.String()
	return out, nil
}

//---------------------------------------------------------------------------------
// procLoop()
// This is the main processing loop that checks for OPA Policies, and updates the
// defined Thunder node as needed.
func procLoop(d axapi.Device, config Configuration) {
	//
	// lookup config.Virts on Thunder to make sure it/they are there.
	vslist, err := d.GetVSlist()
	if err != nil {
		log.Errorf("Error on GetVSlist(): %i\n", err)
	}
	var ff = false
	for _, v := range config.Virts {
		for _, t := range vslist {
			if t.Name == v.Name {
				ff = true
			}
		}
		if !ff {
			log.Errorf("Virtual Server '%s' not found on Thunder node", v.Name)
		}
		ff = false
	}

	//
	// Query OPA with config.THND_ID for BW Policy rate, if needed
	//var bwrate int
	for _, p := range config.Virts {
		if p.Policy == "bw" {
			// --
			// Bandwidth can be controlled on a Thunder node by attaching a "server" template to each server that
			// is assigned to the Service Group that is attached to the Virtual server. This will require two
			// different calls to the Thunder node to retrive first the Service Group name from the Virtual Server,
			// then a call to get the list of 'members' in that Service Group. Then we have to attach the Server Template
			// that we have created with all the bandwidth limitations to each Server. In the end, the 'slb' section will
			// look something like this:
			//
			// slb template server opa-policy-bw
			//   bw-rate-limit 1000 resume 800 duration 20
			// slb server 10.1.1.220 10.1.1.220
			// 	template server opa-policy-bw
			// 	port 31721 tcp
			// slb server 10.1.1.221 10.1.1.221
			// 	template server opa-policy-bw
			// 	port 31721 tcp
			// slb service-group ws-sg tcp
			// 	health-check ws-mon
			// 	member 10.1.1.220 31721
			// 	member 10.1.1.221 31721
			// slb virtual-server ws-vip 10.1.1.44
			// 	port 80 http
			// 		source-nat auto
			// 		service-group ws-sg
			//
			//  Bandwidth Limits are defined as Kbps...so 1000 Kbps = 1 Mbps
			// --
			//
			// Find the policy for the Thunder ID
			opaurl := "http://" + config.OPA_IP + ":" + strconv.Itoa(config.OPA_PORT) + "/v1/data/net/bwnodes/" + config.THND_ID
			out, err := callOPA(opaurl, "GET", "")
			if err != nil {
				log.Warnf("No BW Policy found for Thunder node '%s'\n", config.THND_ID)
			}
			bwpolicy := gjson.Get(out, "result").Array()[0].Str
			if config.Debug > 7 {
				fmt.Printf("policy = %s\n", bwpolicy)
			}
			//
			// Now get the rate
			opaurl = "http://" + config.OPA_IP + ":" + strconv.Itoa(config.OPA_PORT) + "/v1/data/net/bw/" + bwpolicy
			out, err = callOPA(opaurl, "GET", "")
			if err != nil {
				log.Warnf("No Rate found for BW Policy '%s'\n", bwpolicy)
			}
			bwrate := gjson.Get(out, "result").Array()[0].Int()
			if config.Debug > 7 {
				fmt.Printf("rate = %d\n", bwrate)
			}

			//
			// Configure & Set Template on Thunder node for BW Policy
			// NOTE: The BW-Resume var (bwrlr) is hard-coded here at 80% of the BW-Rate collected from the
			// OPA node. This really should be a configuration item.
			// NOTE: The BW-Duration var (bwrld) is hard-coded here for 20 seconds. This really should be a
			// configuration item.
			var resu float32 = 0.8 // This needs to be a config. item -- BW-Resume
			bwrld := 20            // This also needs to be a config. item  -- BW-Duration
			bwrlr := int(float32(bwrate) * resu)
			payload := "{\"server\": {\"name\": \"opa-policy-bw\", \"bw-rate-limit\": " + strconv.Itoa(int(bwrate)) + ", \"bw-rate-limit-resume\": " + strconv.Itoa(bwrlr) + ", \"bw-rate-limit-duration\": " + strconv.Itoa(bwrld) + "} }"
			// -- First, check to see if Template already exists
			out, err = d.GetServerTemplate("opa-policy-bw")
			if err != nil {
				log.Errorf("Error on GetServerTemplate(): %s\n", err)
			}
			if config.Debug > 7 {
				fmt.Println(">>>" + payload)
			}
			if out == "" {
				log.Info("Creating BW Policy Template...")
				err = d.CreateServerTemplate(payload)
				if err != nil {
					log.Errorf("Bandwidth Policy Template could not be created on Thunder node: %s\n", err)
				}
			} else {
				log.Info("Updating BW Policy Template")
				err = d.UpdateServerTemplate(payload)
				if err != nil {
					log.Errorf("Bandwidth Policy Template could not be updated on Thunder node: %s\n", err)
				}
			}
			//
			//  Get Service-Group name & parse out members

			//  Go through list of servers and attach BW Template
		}
		//
		// Query OPA with config.THND_ID for CPS Policy rate, if needed
		if p.Policy == "cps" {
			// --
			// Connection-Rate-Limiting can be configured at an SLB level on a Thunder node by creating a
			// virtual-server Template and attaching it to the SLB. This will limit the Connections-per-Second
			// of the SLB down to the service-group members.  This will require an API call to create the
			// Template, once it has collected the CPS Policy from OPA, and then another API call to attach
			// the Template to the SLB.  Once done, the 'slb' section will look something like this:
			//
			// slb server 10.1.1.220 10.1.1.220
			// 	port 31721 tcp
			// slb server 10.1.1.221 10.1.1.221
			// 	port 31721 tcp
			// slb service-group ws-sg tcp
			// 	health-check ws-mon
			// 	member 10.1.1.220 31721
			// 	member 10.1.1.221 31721
			// slb template virtual-server opa-policy-cps
			//  conn-limit 200
			//  conn-rate-limit 200
			// slb virtual-server ws-vip 10.1.1.44
			//  template virtual-server opa-policy-cps
			// 	port 80 http
			// 		source-nat auto
			// 		service-group ws-sg
			// --
			//
			// Find the policy for the Thunder ID
			opaurl := "http://" + config.OPA_IP + ":" + strconv.Itoa(config.OPA_PORT) + "/v1/data/net/cpsnodes/" + config.THND_ID
			out, err := callOPA(opaurl, "GET", "")
			if err != nil {
				log.Warnf("No CPS Policy found for Thunder node '%s'\n", config.THND_ID)
			}
			cpspolicy := gjson.Get(out, "result").Array()[0].Str
			if config.Debug > 7 {
				fmt.Printf("policy = %s\n", cpspolicy)
			}
			//
			// Now get the rate
			opaurl = "http://" + config.OPA_IP + ":" + strconv.Itoa(config.OPA_PORT) + "/v1/data/net/cps/" + cpspolicy
			out, err = callOPA(opaurl, "GET", "")
			if err != nil {
				log.Warnf("No Rate found for CPS Policy '%s'\n", cpspolicy)
			}
			cpsrate := gjson.Get(out, "result").Array()[0].Int()
			if config.Debug > 7 {
				fmt.Printf("rate = %d\n", cpsrate)
			}

			//
			// Configure & Set Template on Thunder node for CPS Policy
			cpsv := strconv.Itoa(int(cpsrate))
			payload := "{\"virtual-server\": {\"name\": \"opa-policy-cps\", \"conn-limit\": " + cpsv + ", \"conn-rate-limit\": " + cpsv + "} }"
			// -- First, check to see if Template already exists
			out, err = d.GetVirtualServerTemplate("opa-policy-cps")
			if err != nil {
				log.Errorf("Error on GetServerTemplate(): %s\n", err)
			}
			// if not, create, else, update
			if config.Debug > 7 {
				fmt.Println(">>>" + payload)
			}
			if out == "" {
				log.Info("Creating CPS Policy Template...")
				err = d.CreateVirtualServerTemplate(payload)
				if err != nil {
					log.Errorf("CPS Policy Template could not be created on Thunder node: %s\n", err)
				}
			} else {
				log.Info("Updating CPS Policy Template")
				err = d.UpdateVirtualServerTemplate(payload)
				if err != nil {
					log.Errorf("CPS Policy Template could not be updated on Thunder node: %s\n", err)
				}
			}

			//
			// Add Template to SLB
			payload = "{\"virtual-server\": {\"template-virtual-server\": \"opa-policy-cps\" } }"
			err = d.UpdateVirtualServer(p.Name, payload)
			if err != nil {
				log.Errorf("Error updating Virtual Server %s: %s\n", p.Name, err)
			}
		}
	}

	//
	// Configure callbacks/something for changes on OPA?
	// It would be better to implement some sort of check of the OPA Data to see if anything has
	// been changed since the last time this function was run. I can't find anything in the OPA
	// documentation that would allow to retrieve a timestamp or revision number that could be
	// used to determine if indeed a new Data set had been uploaded to OPA, and thus would
	// require a re-run of this function.

}

// RunProcLoop()
//---------------------------------------------------------------------------------
// This fuction just does a simple timed loop based on the config.CHK_INTERVAL number of seconds.
// It does not seem to be the most efficient way to do this, but baring being able to get a
// timestamp or revision number from OPA, we need to do something else to make sure the
// Thunder node is updated for any changes to the OPA Policy & Data.
//
// Future Version?:  Add a .net.revision data leaf to OPA that you would change everytime the OPA
// policy Data is changed. Read it here, and if its different, THEN call procLoop(). This will
// also save lots of log space too, as the Thunder node won't be constantly being updated
// using aXAPIs.
func RunProcLoop(d axapi.Device, config Configuration) {
	// Run forever.....
	interval := time.Second * config.CHK_INTERVAL
	for range time.Tick(interval) {
		procLoop(d, config)
	}
}

//---------------------------------------------------------------------------------
//  MAIN
//---------------------------------------------------------------------------------
func main() {
	//
	// Setup the logging with Timestamps
	customFormat := new(log.TextFormatter)
	customFormat.TimestampFormat = "2006-01-02 15:04:05" // Yes, it MUST be THIS string!
	customFormat.FullTimestamp = true
	log.SetFormatter(customFormat)
	log.Info("A10 Thunder OPA Proxy Starting...")

	//
	// Handle Interrrupts
	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigchan
		ending(2)
	}()

	//
	// Process command line args
	x1 := flag.Int("debug", 0, "Debugging Level")
	x2 := flag.String("opaip", "0.0.0.0", "IP or FQDN of OPA Server")
	x3 := flag.Int("opaport", 8181, "OPA Server API Port")
	x4 := flag.String("thunderip", "0.0.0.0", "IP or FQDN of Thunder node")
	x5 := flag.Int("thunderport", 443, "Thunder node Port")
	x6 := flag.String("thunderid", "", "Thudner node ID")
	x7 := flag.String("config", "./config/config.yaml", "Configuration File Path")
	flag.Parse()
	DEBUG = *x1
	OPA_IP = *x2
	OPA_PORT = *x3
	THND_IP = *x4
	THND_PORT = *x5
	THND_ID = *x6
	CFG_FILE = *x7

	//---------------------------------------------------------------------------------
	// Parse Config File first, then overwrite as needed with Command Line args.
	config, err := getConfig(CFG_FILE)
	if err != nil {
		log.Fatal(err)
	}
	if DEBUG != 0 {
		config.Debug = DEBUG
	}
	if OPA_IP != "0.0.0.0" || config.OPA_IP == "" {
		config.OPA_IP = OPA_IP
	}
	if config.OPA_PORT == 0 {
		config.OPA_PORT = OPA_PORT
	}
	if THND_IP != "0.0.0.0" || config.THND_IP == "" {
		config.THND_IP = THND_IP
	}
	if config.THND_PORT == 0 {
		config.THND_PORT = THND_PORT
	}
	if THND_ID != "" {
		config.THND_ID = THND_ID
	}

	if config.Debug > 7 {
		fmt.Printf("debug: %d\nopaip: %s\nopaport: %d\nthunderip: %s\nthunderport: %d\nthunderid: %s\n",
			config.Debug, config.OPA_IP, config.OPA_PORT, config.THND_IP, config.THND_PORT, config.THND_ID)
	}
	//---------------------------------------------------------------------------------

	//
	// Check for valid args
	ff := 0
	if config.OPA_IP == "0.0.0.0" {
		log.Fatal("Invalid IP address for OPA Server: 0.0.0.0")
		ff = 1
	}
	if config.OPA_PORT == 0 {
		log.Fatal("Invalid Port for OPA Server: 0")
		ff = 1
	}
	if config.THND_IP == "0.0.0.0" {
		log.Fatal("Invalid IP address for Thunder node: 0.0.0.0")
		ff = 1
	}
	if config.THND_PORT == 0 {
		log.Fatal("Invalid Port for Thunder Node: 0")
		ff = 1
	}
	if config.THND_ID == "" {
		log.Fatal("Thunder ID not specified")
		ff = 1
	}
	if ff == 1 {
		// Fatal error, exit program.
		os.Exit(1)
	}

	//
	// Connect to Thunder node
	d := axapi.Device{}
	ap := config.THND_IP + ":" + strconv.Itoa(config.THND_PORT)
	d.Address = ap
	d.Username = config.THND_USER
	d.Password = config.THND_PASSWD
	d, err = d.Login()
	if err != nil {
		log.Fatal(err.Error())
	} else {
		log.Info("Connected to Thunder Device")
	}
	defer d.Logoff()

	//
	// Connect to OPA Server
	//------------------------------------------------------------------------------------------
	// This POC is looking for a specific set of "Policy" defined to the OPA server. Here is the
	// sample Data set that we are using:
	// 	# curl -s http://localhost:30181/v1/data?pretty=true
	// {
	//   "result": {
	//     "net": {
	//       "bw": {
	//         "green": [ "100" ],
	//         "orange": [ "10" ],
	//         "red": [ "0" ],
	//         "yellow": [ "1" ]
	//       },
	//       "bwnodes": {
	//         "thunder-1": [ "orange" ],
	//         "thunder-2": [ "green" ]
	//       },
	//       "cps": {
	//         "blue": [ "1000" ],
	//         "green": [ "10000" ],
	//         "orange": [ "100" ],
	//         "red": [ "0" ],
	//         "yellow": [ "10" ]
	//       },
	//       "cpsnodes": {
	//         "tester1": [ "yellow" ],
	//         "thunder-1": [ "orange" ],
	//         "thunder-2": [ "blue" ]
	//       }
	//     }
	//   }
	// }
	//
	// There are two "Policies" defined in this OPA Data set:  Bandwidth ('bw') and Connection Rate ('cps').
	// For each set there are two sections: one for the actual rates, and the other for the node names and
	// what rate they are to use.
	//
	// If you implment this TCA for all your Thunder nodes, you can control all of them by just changing this
	// Data set and uploading it to OPA. On the next time through the main processing loop, it will adjust the
	// defined SLB virtual-servers with the new Policy values.
	//
	opaurl := "http://" + config.OPA_IP + ":" + strconv.Itoa(config.OPA_PORT) + "/v1/data"
	out, err := callOPA(opaurl, "GET", "")
	if err != nil {
		log.Fatal(err.Error())
		ending(1)
	}

	//
	// Check to see if our 'net' policies are defined
	flg := 0
	if !gjson.Parse(out).Get("result.net.bw").Exists() {
		log.Warn("Bandwidth Plans not found on OPA Server")
		flg++
	}
	if !gjson.Parse(out).Get("result.net.cps").Exists() {
		log.Warn("CPS Plans not found on OPA Server")
		flg++
	}
	if !gjson.Parse(out).Get("result.net.bwnodes").Exists() {
		log.Warn("List of Nodes for Bandwidth Policy not found on OPA Server")
		flg++
	}
	if !gjson.Parse(out).Get("result.net.cpsnodes").Exists() {
		log.Warn("List of Nodes for CPS Policy not found on OPA Server")
		flg++
	}
	// -- We could just error out of the program here if one of the above are not found,
	// but I didn't because there could be other policies that the code implements elsewhere
	// that we want to use.  These are the example policies, and we should check the OPA
	// Server to see if they exist or not if we are going to continue with them.
	if flg == 0 {
		log.Info("Network Policies found on OPA Server")
		// Optional:
		//   ending(1)
	}

	//
	// ** Main processing/policy applying loop
	procLoop(d, config) // Run direct to avoid RunProcLoop() delay on first run.
	RunProcLoop(d, config)

	//---------------------------------------------------------------------------------
	// Exit the Program!
	//ending(0)
}

//---------------------------------------------------------------------------------
// ending() --  Clean up and Exit the program
func ending(val int) {
	// d.Logoff()
	log.Info("A10 Thunder OPA Proxy Ending.")
	os.Exit(val)
}
