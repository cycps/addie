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

	err = InsertSystem("caprica", "root")
	if err != nil {
		t.Error(err)
	}
	_, _, err = SysKey("caprica", "root")
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

func TestOneCreateDestroy(t *testing.T) {

	/*
		err := beginTx()
		if err != nil {
			t.Log(err)
			t.Fatal("failed to start transaction")
		}
	*/

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

	err = InsertSystem("caprica", "root")
	if err != nil {
		t.Log(err)
		t.Fatal("failed to create caprica.root")
	}

	//Add a computer ------------
	c := addie.Computer{}
	//Id
	c.Name = "c"
	c.Sys = "root"
	c.Design = "caprica"
	//NetHost
	c.Interfaces = make(map[string]addie.Interface)
	//Comptuer
	c.Position = addie.Position{0, 0, 0}
	c.OS = "Ubuntu-15.04"
	c.Start_script = "make_muffins.sh"

	err = InsertComputer(c)
	if err != nil {
		t.Log(err)
		t.Fatal("failed to insert computer")
	}

	_c, err := GetComputer(addie.Id{"c", "root", "caprica"})
	if err != nil {
		t.Log(err)
		t.Fatal("failed to retrieve computer")
	}

	if c.Name != _c.Name {
		t.Error("computer round trip failed for: Name")
	}
	if c.Sys != _c.Sys {
		t.Error("computer round trip failed for: Sys")
	}
	if c.Design != _c.Design {
		t.Error("computer round trip failed for: Design")
	}
	//TODO test interfaces
	if c.Position != _c.Position {
		t.Error("computer round trip failed for: Position")
	}
	if c.OS != _c.OS {
		t.Error("computer round trip failed for: OS")
	}
	if c.Start_script != _c.Start_script {
		t.Error("computer round trip failed for: Start_script")
	}

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

	/*
		err = endTx()
		if err != nil {
			t.Log(err)
			t.Fatal("failed to commit transaction")
		}
	*/

}
