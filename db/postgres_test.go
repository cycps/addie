/*
Test suite for Cypress PostgreSQL persistence
*/
package db

import (
	"github.com/cycps/addie"
	"testing"
)

func TestReadDesigns(t *testing.T) {

	designs, err := ReadDesigns()
	if err != nil {
		t.Fatal(err)
	}
	_, ok := designs["design47"]
	if !ok {
		t.Fatal("The database does not contain design47")
	}

}

func TestDesignCreateDestroy(t *testing.T) {

	err := CreateDesign("caprica")
	if err != nil {
		t.Fatal("failed to create caprica")
	}

	designs, err := ReadDesigns()
	if err != nil {
		t.Fatal(err)
	}
	_, ok := designs["caprica"]
	if !ok {
		t.Fatal("caprica has not been created")
	}

	err = DeleteDesign("caprica")
	if err != nil {
		t.Fatal("failed to trash caprica")
	}

	designs, err = ReadDesigns()
	if err != nil {
		t.Fatal(err)
	}
	_, ok = designs["caprica"]
	if ok {
		t.Fatal("caprica persists")
	}

}

func TestSysCreateDestroy(t *testing.T) {

	err := CreateDesign("caprica")
	if err != nil {
		t.Fatal("failed to create caprica")
	}

	designs, err := ReadDesigns()
	if err != nil {
		t.Fatal(err)
	}
	_, ok := designs["caprica"]
	if !ok {
		t.Fatal("caprica has not been created")
	}

	_, err = CreateSystem("caprica", "root")
	if err != nil {
		t.Fatal(err)
	}
	_, err = ReadSysKey("caprica", "root")
	if err != nil {
		t.Fatal(err)
	}

	err = DeleteDesign("caprica")
	if err != nil {
		t.Fatal("failed to trash caprica")
	}

	designs, err = ReadDesigns()
	if err != nil {
		t.Fatal(err)
	}
	_, ok = designs["caprica"]
	if ok {
		t.Fatal("caprica persists")
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

	err := CreateComputer(c)
	if err != nil {
		t.Log(err)
		t.Fatal("failed to insert computer")
	}

	_c, err := ReadComputer(addie.Id{"c", "root", "caprica"})
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

	key, err := ReadIdKey(c.Id)
	if err != nil {
		t.Log(err)
		t.Fatal("computer to modify does not exist")
	}

	d := c
	d.OS = "Kobol"
	d.Name = "b"
	d.Sys = "galactica"
	for k, v := range c.Interfaces {
		d.Interfaces[k] = v
	}

	_, err = UpdateComputer(c.Id, c, d)
	if err != nil {
		t.Log(err)
		t.Fatal("failed to update computer")
	}

	_c, err := ReadComputerByKey(key)
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

func addRouterTest(t *testing.T) addie.Router {

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

	err := CreateRouter(rtr)
	if err != nil {
		t.Log(err)
		t.Fatal("failed to insert router")
	}

	_rtr, err := ReadRouter(addie.Id{"rtr", "root", "caprica"})
	if err != nil {
		t.Log(err)
		t.Fatal("failed to get router")
	}

	if rtr.Id != _rtr.Id {
		t.Fatal("router round trip failed for: id")
	}

	if rtr.PacketConductor != _rtr.PacketConductor {
		t.Fatal("router round trip failed for: packet conductor")
	}

	if rtr.Position != _rtr.Position {
		t.Fatal("router round trip failed for: position")
	}

	return *_rtr
}

func modifyRouterTest(t *testing.T, r addie.Router) addie.Router {

	key, err := ReadIdKey(r.Id)
	if err != nil {
		t.Fatal(err)
	}

	s := r
	s.Name = "bill"
	s.Latency = 222
	s.Capacity = 333

	_, err = UpdateRouter(r.Id, r, s)
	if err != nil {
		t.Fatal(err)
	}

	_r, err := ReadRouterByKey(key)
	if err != nil {
		t.Fatal(err)
	}

	if !s.Equals(_r) {
		t.Log(s)
		t.Log(_r)
		t.Fatal("router update failed")
	}

	return *_r
}

func addSwitchTest(t *testing.T) addie.Switch {

	sw := addie.Switch{}
	//Id
	sw.Name = "sw"
	sw.Sys = "root"
	sw.Design = "caprica"
	//PacketConductor
	sw.Latency = 74
	sw.Capacity = 100000
	sw.Position = addie.Position{100, 100, 20}

	err := CreateSwitch(sw)
	if err != nil {
		t.Log(err)
		t.Fatal("failed to insert switch")
	}

	_sw, err := ReadSwitch(sw.Id)
	if err != nil {
		t.Log(err)
		t.Fatal("failed to get switch")
	}

	if sw.Id != _sw.Id {
		t.Fatal("router round trip failed for: Id")
	}

	if sw.PacketConductor != _sw.PacketConductor {
		t.Fatal("switch round trip failed for: Packet Conductor")
	}
	if sw.Position != _sw.Position {
		t.Fatal("switch round trip failed for: Position")
	}

	return *_sw

}

func modifySwitchTest(t *testing.T, s addie.Switch) addie.Switch {

	key, err := ReadIdKey(s.Id)
	if err != nil {
		t.Fatal(err)
	}

	u := s
	u.Name = "olympic"
	u.Latency = 567
	u.Capacity = 987

	_, err = UpdateSwitch(s.Id, u)
	if err != nil {
		t.Fatal(err)
	}

	_s, err := ReadSwitchByKey(key)
	if err != nil {
		t.Fatal(err)
	}

	if !(u == *_s) {
		t.Log(u)
		t.Log(_s)
		t.Fatal("switch update failed")
	}

	return *_s
}

func addLinkTest(t *testing.T, c addie.Computer) addie.Link {

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

	err := CreateLink(lnk)
	if err != nil {
		t.Log(err)
		t.Fatal("failed to insert link")
	}

	_lnk, err := ReadLink(lnk.Id)
	if err != nil {
		t.Log(err)
		t.Fatal("failed to get link")
	}

	if lnk.Id != _lnk.Id {
		t.Fatal("link round trip failed for: Id")
	}
	if lnk.PacketConductor != _lnk.PacketConductor {
		t.Fatal("link round trip failed for: PacketConductor")
	}
	if lnk.Endpoints != _lnk.Endpoints {
		t.Log("%v != %v", lnk.Endpoints, _lnk.Endpoints)
		t.Fatal("link round trip failed for: Endpoints")
	}

	return lnk

}

func modifyLinkTest(t *testing.T, l addie.Link) addie.Link {
	key, err := ReadIdKey(l.Id)
	if err != nil {
		t.Fatal(err)
	}

	m := l
	m.Name = "helo"
	m.Latency = 1234
	m.Capacity = 4680

	_, err = UpdateLink(l.Id, m)
	if err != nil {
		t.Fatal(err)
	}

	_l, err := ReadLinkByKey(key)
	if err != nil {
		t.Fatal(err)
	}

	if !(m.Equals(_l)) {
		t.Log(m)
		t.Log(_l)
		t.Fatal("link update failed")
	}

	return *_l

}

func TestOneCreateDestroyUpdate(t *testing.T) {

	err := CreateDesign("caprica")
	if err != nil {
		t.Log(err)
		t.Fatal("failed to create caprica")
	}

	//ghetto transaction
	defer func() {
		err = DeleteDesign("caprica") //on cascade delete cleans up everything
		if err != nil {
			t.Log(err)
			t.Fatal("failed to trash caprica")
		}
	}()

	_, err = CreateSystem("caprica", "root")
	if err != nil {
		t.Log(err)
		t.Fatal("failed to create caprica.root")
	}

	c := addComputerTest(t)
	c = modifyComputerTest(t, c)

	r := addRouterTest(t)
	modifyRouterTest(t, r)

	s := addSwitchTest(t)
	modifySwitchTest(t, s)

	l := addLinkTest(t, c)
	modifyLinkTest(t, l)
}
