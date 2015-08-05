/*
Test suite for Cypress PostgreSQL persistence
*/
package db

import (
	"github.com/cycps/addie"
	"testing"
)

func TestGetDesigns(t *testing.T) {

	designs, err := GetDesigns()
	if err != nil {
		t.Error(err)
	}
	_, ok := designs["design47"]
	if !ok {
		t.Error("The database does not contain design47")
	}

}

func TestDesignCreateDestroy(t *testing.T) {

	err := InsertDesign("caprica")
	if err != nil {
		t.Error("failed to create caprica")
	}

	designs, err := GetDesigns()
	if err != nil {
		t.Error(err)
	}
	_, ok := designs["caprica"]
	if !ok {
		t.Error("caprica has not been created")
	}

	err = TrashDesign("caprica")
	if err != nil {
		t.Error("failed to trash caprica")
	}

	designs, err = GetDesigns()
	if err != nil {
		t.Error(err)
	}
	_, ok = designs["caprica"]
	if ok {
		t.Error("caprica persists")
	}

}

func TestSysCreateDestroy(t *testing.T) {

	err := InsertDesign("caprica")
	if err != nil {
		t.Error("failed to create caprica")
	}

	designs, err := GetDesigns()
	if err != nil {
		t.Error(err)
	}
	_, ok := designs["caprica"]
	if !ok {
		t.Error("caprica has not been created")
	}

	_, err = InsertSystem("caprica", "root")
	if err != nil {
		t.Error(err)
	}
	_, err = SysKey("caprica", "root")
	if err != nil {
		t.Error(err)
	}

	err = TrashDesign("caprica")
	if err != nil {
		t.Error("failed to trash caprica")
	}

	designs, err = GetDesigns()
	if err != nil {
		t.Error(err)
	}
	_, ok = designs["caprica"]
	if ok {
		t.Error("caprica persists")
	}

}

func addComputerTest(t *testing.T) addie.Computer {

	c := addie.Computer{}
	//Id
	c.Name = "c"
	c.Sys = "root"
	c.Design = "caprica"
	//NetHost
	c.Interfaces = make(map[string]addie.Interface)
	ifx := addie.Interface{}
	ifx.Name = "eth0"
	ifx.Capacity = 345
	ifx.Latency = 2
	c.Interfaces["eth0"] = ifx
	//Comptuer
	c.Position = addie.Position{0, 0, 0}
	c.OS = "Ubuntu-15.04"
	c.Start_script = "make_muffins.sh"

	err := InsertComputer(c)
	if err != nil {
		t.Log(err)
		t.Fatal("failed to insert computer")
	}

	_c, err := GetComputer(addie.Id{"c", "root", "caprica"})
	if err != nil {
		t.Log(err)
		t.Fatal("failed to retrieve computer")
	}

	if !c.Equals(_c) {
		t.Fatal("computer round trip failed")
	}

	return c

}

func modifyComputerTest(t *testing.T, c addie.Computer) addie.Computer {

	key, err := IdKey(c.Id)
	if err != nil {
		t.Log(err)
		t.Fatal("computer to modify does not exist")
	}

	d := c
	d.OS = "Kobol"
	d.Name = "b"
	d.Sys = "galactica"

	err = UpdateComputer(c.Id, d)
	if err != nil {
		t.Log(err)
		t.Fatal("failed to update computer")
	}

	_c, err := GetComputerByKey(key)
	if err != nil {
		t.Log(err)
		t.Fatal("failed to retrieve updated computer")
	}

	if !d.Equals(_c) {
		t.Log(c)
		t.Log(_c)
		t.Fatal("computer update check failed")
	}

	return *_c
}

func TestOneCreateDestroyUpdate(t *testing.T) {

	err := InsertDesign("caprica")
	if err != nil {
		t.Log(err)
		t.Fatal("failed to create caprica")
	}

	//ghetto transaction
	defer func() {
		err = TrashDesign("caprica") //on cascade delete cleans up everything
		if err != nil {
			t.Log(err)
			t.Fatal("failed to trash caprica")
		}
	}()

	_, err = InsertSystem("caprica", "root")
	if err != nil {
		t.Log(err)
		t.Fatal("failed to create caprica.root")
	}

	c := addComputerTest(t)
	c = modifyComputerTest(t, c)

	//Add a router ------------
	rtr := addie.Router{}
	//Id
	rtr.Name = "rtr"
	rtr.Sys = "root"
	rtr.Design = "caprica"
	//Net Host
	rtr.Interfaces = make(map[string]addie.Interface)
	//PacketConductor
	rtr.Latency = 47
	rtr.Capacity = 1000
	rtr.Position = addie.Position{4, 7, 4}

	err = InsertRouter(rtr)
	if err != nil {
		t.Log(err)
		t.Fatal("failed to insert router")
	}

	_rtr, err := GetRouter(addie.Id{"rtr", "root", "caprica"})
	if err != nil {
		t.Log(err)
		t.Fatal("failed to get router")
	}

	if rtr.Id != _rtr.Id {
		t.Error("router round trip failed for: id")
	}

	if rtr.PacketConductor != _rtr.PacketConductor {
		t.Error("router round trip failed for: packet conductor")
	}

	if rtr.Position != _rtr.Position {
		t.Error("router round trip failed for: position")
	}

	//Add a switch ---------------
	sw := addie.Switch{}
	//Id
	sw.Name = "sw"
	sw.Sys = "root"
	sw.Design = "caprica"
	//PacketConductor
	sw.Latency = 74
	sw.Capacity = 100000
	sw.Position = addie.Position{100, 100, 20}

	err = InsertSwitch(sw)
	if err != nil {
		t.Log(err)
		t.Fatal("failed to insert switch")
	}

	_sw, err := GetSwitch(sw.Id)
	if err != nil {
		t.Log(err)
		t.Fatal("failed to get switch")
	}

	if sw.Id != _sw.Id {
		t.Error("router round trip failed for: Id")
	}

	if sw.PacketConductor != _sw.PacketConductor {
		t.Error("switch round trip failed for: Packet Conductor")
	}
	if sw.Position != _sw.Position {
		t.Error("switch round trip failed for: Position")
	}

	//Add a link -----------------
	lnk := addie.Link{}
	//id
	lnk.Name = "lnk"
	lnk.Sys = "root"
	lnk.Design = "caprica"
	//packet conductor
	lnk.Latency = 123
	lnk.Capacity = 2468
	//endpoints
	lnk.Endpoints[0] = addie.NetIfRef{c.Id, "eth0"}
	lnk.Endpoints[1] = addie.NetIfRef{c.Id, "eth0"}

	err = InsertLink(lnk)
	if err != nil {
		t.Log(err)
		t.Fatal("failed to insert link")
	}

	_lnk, err := GetLink(lnk.Id)
	if err != nil {
		t.Log(err)
		t.Fatal("failed to get link")
	}

	if lnk.Id != _lnk.Id {
		t.Error("link round trip failed for: Id")
	}
	if lnk.PacketConductor != _lnk.PacketConductor {
		t.Error("link round trip failed for: PacketConductor")
	}
	if lnk.Endpoints != _lnk.Endpoints {
		t.Log("%v != %v", lnk.Endpoints, _lnk.Endpoints)
		t.Error("link round trip failed for: Endpoints")
	}

}
