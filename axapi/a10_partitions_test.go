//
//  a10_partitions.go tests
//

package axapi

import (
	"fmt"
	"strconv"
	"testing"
)

func TestGetPartitionList(t *testing.T) {
	d := setup()
	p, err := d.GetPartitionList()
	notErr(t, err)
	fmt.Print("Current Partitions = ")
	fmt.Println(p)
}

func TestGetAvailablePartitionIDs(t *testing.T) {
	d := setup()
	f, err := d.GetAvailablePartitionIDs()
	notErr(t, err)
	fmt.Println(f)
}

func TestCreatePartition(t *testing.T) {
	name := "btest"
	d := setup()
	i, err := d.CreatePartition(name, "cgnv6")
	notErr(t, err)
	fmt.Println("New Partition ID = " + strconv.Itoa(i))

	err = d.DeletePartition(name)
	notErr(t, err)
}

func TestGetActivePartition(t *testing.T) {
	d := setup()
	f, err := d.GetActivePartition()
	notErr(t, err)
	assert(t, f, "shared")
}

func TestGetMaxPartitions(t *testing.T) {
	d := setup()
	f, err := d.GetMaxPartitions()
	notErr(t, err)
	assertNot(t, f, 0)
}
