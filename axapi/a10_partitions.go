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
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

// Partition is used to hold the partition data
type Partition struct {
	Name       string `json:"partition-name"`
	ID         int    `json:"partition-id"`
	Type       string `json:"partition-type"`
	Parent     string `json:"parent-l3v"`
	AppType    string `json:"application-Type"`
	AdminCount int    `json:"admin-Count"`
	Status     string `json:"status"`
}

// Oper is a placeholder for a JSON struct
type Oper struct {
	Partitions []Partition `json:"partition-list"`
}

// PartAll is a placeholder for a JSON struct
type PartAll struct {
	Oper Oper `json:"oper"`
}

// PartitionList holds a list of all Active & Non-Active Partitions defined on the Thunder device
type PartitionList struct {
	All PartAll `json:"partition-all"`
}

// GetPartition - Get the info on the given partition
//-----------------------------------------------------------------------------
func (d Device) GetPartition(name string) (Partition, error) {
	p := new(Partition)
	v, err := d.GetPartitionList()
	if err != nil {
		return Partition{}, err
	}
	for i := range v.All.Oper.Partitions {
		if v.All.Oper.Partitions[i].Name == name {
			p = &v.All.Oper.Partitions[i]
			return *p, nil
		}
	}

	return Partition{}, errors.New("Partition Name not found")
}

// GetPartitionList - Get the list of partitions from the Thunder Device
//-----------------------------------------------------------------------------
func (d Device) GetPartitionList() (PartitionList, error) {
	p := PartitionList{}
	body, err := _restCall(d, "/partition-all/oper", "GET", nil)
	if err != nil {
		return PartitionList{}, err
	}

	//fmt.Println(string(body))
	err = json.Unmarshal(body, &p)
	if err != nil {
		return PartitionList{}, err
	}
	return p, nil
}

// GetAvailablePartitionIDs - grab the available id array(s)
//-----------------------------------------------------------------------------
func (d Device) GetAvailablePartitionIDs() ([]int, error) {
	var rr []int
	body, err := _restCall(d, "/partition-available-id/oper", "GET", nil)
	if err != nil {
		return []int{}, err
	}
	g := int(gjson.GetBytes(body, "partition-available-id.oper.range-list.#").Int())
	for i := 0; i < g; i++ {
		t1 := "partition-available-id.oper.range-list." + strconv.Itoa(i) + ".start"
		t2 := "partition-available-id.oper.range-list." + strconv.Itoa(i) + ".end"
		s, _ := strconv.Atoi(gjson.GetBytes(body, t1).Str)
		e, _ := strconv.Atoi(gjson.GetBytes(body, t2).Str)
		for k := s; k <= e; k++ {
			rr = append(rr, k)
		}
	}

	return rr, nil
}

// CreatePartition - Make a new Partition on the Thunder Device
//-----------------------------------------------------------------------------
func (d Device) CreatePartition(name string, t string) (int, error) {
	// Get the next available partition number
	ids, err := d.GetAvailablePartitionIDs()
	if err != nil {
		return 0, err
	}
	id := ids[0] // first available id will be in [0]

	payload := strings.NewReader(" ")
	if t != "" {
		payload = strings.NewReader("{ \"partition\": {\"partition-name\": \"" + name + "\", \"id\": " + strconv.Itoa(id) + ", \"application-type\": \"" + t + "\"} }")
	} else {
		payload = strings.NewReader("{ \"partition\": {\"partition-name\": \"" + name + "\", \"id\": \"" + strconv.Itoa(id) + "\"} }")
	}

	//fmt.Println(payload)
	body, err := _restCall(d, "/partition", "POST", payload)
	if err != nil {
		return 0, err
	}
	if e, msg := d.chkResp(body); e {
		return 0, msg
	}

	//fmt.Println(string(body))

	return id, nil
}

// DeletePartition - remove the partition from the Thunder device
//-----------------------------------------------------------------------------
func (d Device) DeletePartition(name string) error {
	// lookup partition to make sure its there.
	g, err := d.GetPartition(name)
	if err != nil {
		return err
	}
	id := g.ID

	// First, "delete" the partition (IE> Mark it as "Not-Active")
	url := "/partition/" + name
	body, err := _restCall(d, url, "DELETE", nil)
	if err != nil {
		return err
	}

	if e, msg := d.chkResp(body); e {
		return msg
	}

	// Them, Remove the partition from the Thunder device
	//payload := strings.NewReader(" ")
	payload := strings.NewReader("{ \"partition\": {\"partition-name\": \"" + name + "\", \"id\": \"" + strconv.Itoa(id) + "\"} }")

	body, err = _restCall(d, "/delete/partition/", "POST", payload)
	if err != nil {
		return err
	}

	if e, msg := d.chkResp(body); e {
		return msg
	}

	return nil

}

// GetActivePartition - What is the current Active Partition?
//-----------------------------------------------------------------------------
func (d Device) GetActivePartition() (string, error) {
	body, err := _restCall(d, "/active-partition", "GET", nil)
	if err != nil {
		return "", err
	}

	if e, msg := d.chkResp(body); e {
		return "", msg
	}

	return gjson.GetBytes(body, "active-partition.partition-name").Str, nil
}

// GetMaxPartitions - How many Partitions can this Thunder Device support?
//-----------------------------------------------------------------------------
func (d Device) GetMaxPartitions() (int, error) {
	body, err := _restCall(d, "/techreport/max-partitions", "GET", nil)
	if err != nil {
		return 0, err
	}

	if e, msg := d.chkResp(body); e {
		return 0, msg
	}

	return int(gjson.GetBytes(body, "max-partitions.value").Int()), nil
}
