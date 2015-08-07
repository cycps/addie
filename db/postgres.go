/*
This file contains the code for persisting Cypress data model elements to
PostgreSQL
*/
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/cycps/addie"
	_ "github.com/lib/pq"
	"log"
	"path"
	"runtime"
	"strings"
)

//Common Variables ------------------------------------------------------------

const (
	dbAddr = "192.168.1.201"
)

var db *sql.DB = nil
var tx *sql.Tx = nil

func dbConnect() error {
	var err error = nil
	if db == nil {
		log.Println("dbConnect")
		db, err = sql.Open("postgres", "host="+dbAddr+" user=root dbname=cyp")
	}
	if err != nil {
		log.Println(err)
		return errors.New("Could not open DB connection")
	}
	//db.SetMaxOpenConns(5)
	return nil
}

//Common Functions ------------------------------------------------------------

func dbPing() error {

	if db == nil {
		err := dbConnect()
		if err != nil {
			return err
		}
	}

	err := db.Ping()
	if err != nil {
		db = nil
		log.Println("dbPing failed -- trying to reconnect")
		log.Println(err)

		err := dbConnect()
		if err != nil {
			log.Println(err)
			return errors.New("could not ping DB")
		}

		return nil
	}
	return nil
}

func RunQ(q string) (*sql.Rows, error) {
	return runQ(q)
}
func runQ(q string) (*sql.Rows, error) {
	err := dbPing()
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to run query -- error communicating with DB")
	}
	var rows *sql.Rows
	if tx != nil {
		rows, err = tx.Query(q)
	} else {
		rows, err = db.Query(q)
	}
	if err != nil {
		log.Println(err)
		return nil, errors.New("Failed to run query")
	}
	return rows, nil
}

func safeClose(rows *sql.Rows) {
	if rows != nil {
		rows.Close()
	}
}

func runC(q string) error {
	rows, err := runQ(q)
	safeClose(rows)
	return err
}

func beginTx() error {
	err := dbPing()
	if err != nil {
		log.Println(err)
		return errors.New("failed to start transaction -- error communicating with DB")
	}
	tx, err = db.Begin()
	if err != nil {
		log.Println(err)
		return errors.New("Failed to start transaction")
	}
	return nil
}

func endTx() error {
	err := dbPing()
	if err != nil {
		log.Println(err)
		return errors.New("failed to commit transaction -- error communicating with DB")
	}
	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return errors.New("Failed to commit transaction")
	}
	tx = nil
	return nil
}

func getKey(q string) (int, error) {

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		log.Println(err)
		return -1, fmt.Errorf("failed to run key query: %s", q)
	}

	if !rows.Next() {
		return -1, fmt.Errorf("could not find key")
	}

	var sys_id int
	err = rows.Scan(&sys_id)
	if err != nil {
		log.Println(err)
		return -1, fmt.Errorf("could not read key")
	}

	return sys_id, nil

}

// Failure functions -----------------------------------------------------------

func callerFailure(cause error, msg string) error {

	pc := make([]uintptr, 1)
	runtime.Callers(3, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	fn := strings.TrimPrefix(f.Name(), "github.com/cycps/addie/db.")

	log.Println(cause)
	return fmt.Errorf("%s:%d [%s] %s", path.Base(file), line, fn, msg)

}

func insertFailure(cause error) error {
	return callerFailure(cause, "insert failed")
}

func selectFailure(cause error) error {
	return callerFailure(cause, "select failed")
}

func scanFailure(cause error) error {
	return callerFailure(cause, "result scan failed")
}

func deleteFailure(cause error) error {
	return callerFailure(cause, "delete failed")
}

func readFailure(cause error) error {
	return callerFailure(cause, "read failed")
}

func updateFailure(cause error) error {
	return callerFailure(cause, "update failed")
}

func createFailure(cause error) error {
	return callerFailure(cause, "create failed")
}

func transactBeginFailure(cause error) error {
	return callerFailure(cause, "transaction begin failed")
}

func transactEndFailure(cause error) error {
	return callerFailure(cause, "transaction end failed")
}

func emptyReadbackFailure() error {
	return callerFailure(fmt.Errorf("the insert readback resulted in 0 rows"),
		"empty readback failure")
}

func emptyReadFailure() error {
	return callerFailure(fmt.Errorf("the query resulted in 0 rows"),
		"empty query failure")
}

func notImplementedFailure() error {
	return callerFailure(fmt.Errorf("not implemented"), "not implemented failure")
}

// CRUD ========================================================================

// Users -----------------------------------------------------------------------

func ReadUserKey(name string) (int, error) {

	q := fmt.Sprintf("SELECT id FROM users WHERE name = '%s'", name)
	key, err := getKey(q)
	if err != nil {
		return -1, selectFailure(err)
	}

	return key, nil
}

// Designs ---------------------------------------------------------------------

func CreateDesign(name string, user string) error {

	//grab the user key
	user_key, err := ReadUserKey(user)
	if err != nil {
		return readFailure(err)
	}

	q := fmt.Sprintf("INSERT INTO designs (name, owner) VALUES ('%s', %d)",
		name, user_key)

	err = runC(q)
	if err != nil {
		return insertFailure(err)
	}

	//every design starts life with a 'root' system
	_, err = CreateSystem(name, "root")
	if err != nil {
		return createFailure(err)
	}

	return nil

}

func ReadDesigns() (map[string]struct{}, error) {

	q := "SELECT (name) FROM designs"
	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		return nil, selectFailure(err)
	}

	//Place holder empty struct, designs may hold more data in the future
	var e struct{}
	//map to hold the result
	m := make(map[string]struct{})

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return nil, scanFailure(err)
		}
		m[name] = e
	}

	return m, nil

}

func ReadDesignKey(name string) (int, error) {

	q := fmt.Sprintf("SELECT id FROM designs WHERE name = '%s'", name)
	key, err := getKey(q)
	if err != nil {
		return -1, selectFailure(err)
	}
	return key, nil

}

func DeleteDesign(name string) error {

	q := fmt.Sprintf("DELETE FROM designs WHERE name = '%s'", name)
	err := runC(q)
	if err != nil {
		return deleteFailure(err)
	}

	return nil
}

func ReadDesign(name string) (*addie.Design, error) {

	/*
		err := beginTx()
		if err != nil {
			return nil, transactBeginFailure(err)
		}
	*/

	key, err := ReadDesignKey(name)
	if err != nil {
		return nil, readFailure(err)
	}

	dsg := addie.EmptyDesign(name)

	//we are going to go from the top down, grabbing all of the systems
	//and then grabbing the components of the systems
	q := fmt.Sprintf("SELECT id FROM systems WHERE design_id = %d", key)

	rows, err := runQ(q)
	defer safeClose(rows)

	if err != nil {
		return nil, selectFailure(err)
	}

	for rows.Next() {

		var sys_key int
		err := rows.Scan(&sys_key)
		if err != nil {
			return nil, scanFailure(err)
		}

		//computers
		computers, err := ReadSystemComputers(sys_key)
		if err != nil {
			return nil, readFailure(err)
		}
		for _, c := range computers {
			dsg.Elements[c.Id] = c
		}

		//switches
		switches, err := ReadSystemSwitches(sys_key)
		if err != nil {
			return nil, readFailure(err)
		}
		for _, s := range switches {
			dsg.Elements[s.Id] = s
		}

		//routers
		routers, err := ReadSystemRouters(sys_key)
		if err != nil {
			return nil, readFailure(err)
		}
		for _, r := range routers {
			dsg.Elements[r.Id] = r
		}

		//links
		links, err := ReadSystemLinks(sys_key)
		if err != nil {
			return nil, readFailure(err)
		}
		for _, l := range links {
			dsg.Elements[l.Id] = l
		}

	}

	/*
		endTx()
		if err != nil {
			return nil, transactEndFailure(err)
		}
	*/

	return &dsg, nil
}

// Systems --------------------------------------------------------------------

func CreateSystem(design string, name string) (int, error) {

	design_key, err := ReadDesignKey(design)
	if err != nil {
		return -1, readFailure(err)
	}

	q := fmt.Sprintf(
		"INSERT INTO systems (design_id, name) VALUES (%d, '%s') RETURNING id",
		design_key, name)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		return -1, insertFailure(err)
	}
	if !rows.Next() {
		return -1, emptyReadbackFailure()
	}
	var id_key int
	err = rows.Scan(&id_key)
	if err != nil {
		return -1, scanFailure(err)
	}

	return id_key, nil

}

func ReadSysKey(design string, name string) (int, error) {

	design_key, err := ReadDesignKey(design)
	if err != nil {
		return -1, readFailure(err)
	}

	q := fmt.Sprintf("SELECT id FROM systems WHERE name = '%s' AND design_id = %d",
		name, design_key)

	sys_key, err := getKey(q)
	if err != nil {
		return -1, selectFailure(err)
	}

	return sys_key, nil

}

func ReadSystemComputers(key int) ([]addie.Computer, error) {

	var result []addie.Computer

	q := fmt.Sprintf(
		"SELECT computers.id FROM computers "+
			"INNER JOIN ids on computers.id = ids.id "+
			"WHERE ids.sys_id = %d", key)

	rows, err := runQ(q)
	defer safeClose(rows)

	if err != nil {
		return nil, selectFailure(err)
	}

	for rows.Next() {
		var comp_key int
		err := rows.Scan(&comp_key)
		if err != nil {
			return nil, scanFailure(err)
		}

		comp, err := ReadComputerByKey(comp_key)
		if err != nil {
			return nil, readFailure(err)
		}
		result = append(result, *comp)
	}

	return result, nil
}

func SysRecycle() error {

	//TODO

	return nil

}

// Ids ------------------------------------------------------------------------

func ReadIdKey(id addie.Id) (int, error) {

	sys_key, err := ReadSysKey(id.Design, id.Sys)
	if err != nil {
		return -1, readFailure(err)
	}

	q := fmt.Sprintf("SELECT id FROM ids WHERE name = '%s' AND sys_id = '%d'",
		id.Name, sys_key)

	id_key, err := getKey(q)
	if err != nil {
		return -1, selectFailure(err)
	}
	return id_key, nil
}

func CreateId(id addie.Id) (int, error) {
	sys_id, err := ReadSysKey(id.Design, id.Sys)
	if err != nil {
		return -1, readFailure(err)
	}

	q := fmt.Sprintf(
		"INSERT INTO ids (name, sys_id) VALUES ('%s', %d) RETURNING id",
		id.Name, sys_id)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		return -1, insertFailure(err)
	}

	if !rows.Next() {
		return -1, emptyReadbackFailure()
	}
	var id_key int
	err = rows.Scan(&id_key)
	if err != nil {
		return -1, scanFailure(err)
	}

	return id_key, nil
}

func ReadId(id int) (*addie.Id, error) {

	_id := new(addie.Id)

	q := fmt.Sprintf(
		"select ids.name, systems.name, designs.name "+
			"from ids "+
			"inner join systems on ids.sys_id = systems.id "+
			"inner join designs on systems.design_id = designs.id "+
			"where ids.id = %d", id)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		return nil, selectFailure(err)
	}
	if !rows.Next() {
		return nil, emptyReadFailure()
	}
	err = rows.Scan(&_id.Name, &_id.Sys, &_id.Design)
	if err != nil {
		return nil, scanFailure(err)
	}

	return _id, nil

}

/*
UpdateID updates an id. If the system in the new id does not exist it is created.
Changing design is not supported through this interface
*/
func UpdateId(oid addie.Id, id addie.Id) (int, error) {

	key, err := ReadIdKey(oid)
	if err != nil {
		return -1, readFailure(err)
	}

	if oid == id {
		return key, nil
	}
	if oid.Design != id.Design {
		return key,
			fmt.Errorf("[UpdateId] changing design though this interface not supported")
	}

	if oid.Sys != id.Sys {
		sys_key, err := ReadSysKey(id.Design, id.Sys)
		if err != nil {
			sys_key, err = CreateSystem(id.Design, id.Sys)
			if err != nil {
				return key, createFailure(err)
			}
		}
		q := fmt.Sprintf(
			"UPDATE ids SET sys_id = %d WHERE id = %d", sys_key, key)

		err = runC(q)
		if err != nil {
			return key, updateFailure(err)
		}

		//err := SysRecycle() do this in background?
		//if err != nil {
		//	log.Println(err)
		//	return key, fmt.Errorf("[UpdateId] an error occured during recycling")
		//}
	}
	if oid.Name != id.Name {
		q := fmt.Sprintf(
			"UPDATE ids SET name = '%s' WHERE id = %d", id.Name, key)
		err = runC(q)
		if err != nil {
			return key, updateFailure(err)
		}
	}

	return key, nil
}

// Interfaces ------------------------------------------------------------------------

func ReadInterfaceKey(host_id int, ifname string) (int, error) {

	q := fmt.Sprintf(
		"SELECT id FROM interfaces WHERE host_id = %d AND name = '%s'",
		host_id, ifname)

	key, err := getKey(q)
	if err != nil {
		log.Printf("hid=%d, if=%s", host_id, ifname)
		return -1, selectFailure(err)
	}

	return key, nil

}

func CreateInterface(host_id int, ifx addie.Interface) error {

	pkt_id, err := CreatePacketConductor(ifx.PacketConductor)
	if err != nil {
		return createFailure(err)
	}

	q := fmt.Sprintf(
		"INSERT INTO interfaces (name, host_id, packet_conductor_id) "+
			"VALUES ('%s', %d, %d)",
		ifx.Name, host_id, pkt_id)

	err = runC(q)
	if err != nil {
		return insertFailure(err)
	}

	return nil
}

func UpdateInterface(host_id int, old addie.Interface, ifx addie.Interface) error {

	q := fmt.Sprintf(
		"SELECT id, packet_conductor_id FROM interfaces "+
			"WHERE host_id = %d AND name = '%s'",
		host_id, old.Name)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		return selectFailure(err)
	}
	var key, pkt_key int
	err = rows.Scan(&key, &pkt_key)
	if err != nil {
		return scanFailure(err)
	}

	_, err = UpdatePacketConductor(pkt_key, ifx.PacketConductor)
	if err != nil {
		return updateFailure(err)
	}

	q = fmt.Sprintf(
		"UPDATE interfaces SET name ='%s' WHERE id = %d", ifx.Name, key)
	err = runC(q)
	if err != nil {
		return updateFailure(err)
	}

	return nil
}

func ReadInterface(id int) (*addie.Interface, error) {

	q := fmt.Sprintf(
		"SELECT name, host_id, packet_conductor_id FROM interfaces WHERE id = %d", id)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		return nil, selectFailure(err)
	}
	if !rows.Next() {
		return nil, emptyReadFailure()
	}
	var name string
	var host_key, pkt_key int
	err = rows.Scan(&name, &host_key, &pkt_key)
	if err != nil {
		return nil, scanFailure(err)
	}

	pkt, err := ReadPacketConductor(pkt_key)
	if err != nil {
		return nil, readFailure(err)
	}

	ifx := addie.Interface{}
	ifx.Name = name
	ifx.PacketConductor = *pkt

	return &ifx, nil

}

func ReadHostInterfaces(host_id int) (*map[string]addie.Interface, error) {

	q := fmt.Sprintf(
		"SELECT name, packet_conductor_id FROM interfaces WHERE host_id = %d", host_id)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		return nil, selectFailure(err)
	}
	result := make(map[string]addie.Interface)
	for rows.Next() {
		var name string
		var pkt_key int
		err = rows.Scan(&name, &pkt_key)
		if err != nil {
			return nil, scanFailure(err)
		}

		pkt, err := ReadPacketConductor(pkt_key)
		if err != nil {
			return nil, readFailure(err)
		}

		ifx := addie.Interface{}
		ifx.Name = name
		ifx.PacketConductor = *pkt

		result[name] = ifx

	}

	return &result, nil
}

// Network Hosts ---------------------------------------------------------------------

func CreateNetworkHost(h addie.NetHost) (int, error) {

	//id insert
	id_key, err := CreateId(h.Id)
	if err != nil {
		return -1, createFailure(err)
	}

	//network host insert
	err = CreateNetworkHostByKey(id_key)
	if err != nil {
		return -1, createFailure(err)
	}

	return id_key, nil
}

func CreateNetworkHostByKey(id_key int) error {

	q := fmt.Sprintf(
		"INSERT INTO network_hosts (id) VALUES (%d)", id_key)

	err := runC(q)
	if err != nil {
		return insertFailure(err)
	}

	return nil

}

func UpdateNetworkHost(oid addie.Id, old addie.NetHost, h addie.NetHost) (int, error) {

	key, err := UpdateId(oid, h.Id)
	if err != nil {
		return -1, updateFailure(err)
	}

	for k, v := range h.Interfaces {
		_v, ok := old.Interfaces[k]
		if ok && _v == v {
			//log.Printf("ifx %v == %v", _v, v)
			continue
		} else if ok && _v != v {
			//log.Printf("ifx % %v --> %v", _v, v)
			UpdateInterface(key, _v, v)
		} else {
			//log.Printf("ifx + %v", v)
			CreateInterface(key, v)
		}
	}

	return key, nil

}

// Positions -------------------------------------------------------------------------

func CreatePosition(p addie.Position) (int, error) {

	q := fmt.Sprintf(
		"INSERT INTO positions (x, y, z) VALUES (%f, %f, %f) RETURNING id",
		p.X, p.Y, p.Z)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		return -1, insertFailure(err)
	}
	if !rows.Next() {
		return -1, emptyReadbackFailure()
	}

	var pos_key int
	err = rows.Scan(&pos_key)
	if err != nil {
		return -1, scanFailure(err)
	}

	return pos_key, nil
}

func ReadPosition(id int) (*addie.Position, error) {

	q := fmt.Sprintf(
		"SELECT x, y, z FROM positions WHERE id = %d", id)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		return nil, selectFailure(err)
	}

	if !rows.Next() {
		return nil, emptyReadFailure()
	}

	var x, y, z float32
	err = rows.Scan(&x, &y, &z)
	if err != nil {
		return nil, scanFailure(err)
	}

	return &addie.Position{x, y, z}, nil

}

func UpdatePosition(key int, p addie.Position) (int, error) {

	q := fmt.Sprintf("UPDATE positions SET x = %f, y = %f, z = %f WHERE id = %d",
		p.X, p.Y, p.Z, key)

	err := runC(q)
	if err != nil {
		return key, updateFailure(err)
	}

	return key, nil
}

// Packet Conductors -----------------------------------------------------------------

func CreatePacketConductor(p addie.PacketConductor) (int, error) {

	q := fmt.Sprintf(
		"INSERT INTO packet_conductors (capacity, latency) VALUES (%d, %d) "+
			"RETURNING id",
		p.Capacity, p.Latency)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		return -1, insertFailure(err)
	}
	if !rows.Next() {
		return -1, emptyReadbackFailure()
	}

	var pkt_key int
	err = rows.Scan(&pkt_key)
	if err != nil {
		return -1, scanFailure(err)
	}

	return pkt_key, nil

}

func ReadPacketConductor(id int) (*addie.PacketConductor, error) {

	q := fmt.Sprintf(
		"SELECT capacity, latency FROM packet_conductors WHERE id = %d", id)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		return nil, selectFailure(err)
	}
	if !rows.Next() {
		return nil, emptyReadFailure()
	}

	var capacity, latency int
	err = rows.Scan(&capacity, &latency)
	if err != nil {
		return nil, scanFailure(err)
	}

	return &addie.PacketConductor{capacity, latency}, nil

}

func UpdatePacketConductor(key int, p addie.PacketConductor) (int, error) {

	q := fmt.Sprintf(
		"UPDATE packet_conductors SET capacity = %d, latency = %d WHERE id = %d",
		p.Capacity, p.Latency, key)

	err := runC(q)
	if err != nil {
		return key, updateFailure(err)
	}

	return key, nil

}

// Computers -------------------------------------------------------------------------

func CreateComputer(c addie.Computer) error {

	//nethost insert
	host_key, err := CreateNetworkHost(c.NetHost)
	if err != nil {
		return createFailure(err)
	}

	//position insert
	pos_key, err := CreatePosition(c.Position)
	if err != nil {
		return createFailure(err)
	}

	//interfaces insert
	for _, ifx := range c.Interfaces {
		err = CreateInterface(host_key, ifx)
		if err != nil {
			return createFailure(err)
		}
	}

	//computer insert
	q := fmt.Sprintf("INSERT INTO computers (id, os, start_script, position_id) "+
		"values (%d, '%s', '%s', %d)", host_key, c.OS, c.Start_script, pos_key)

	err = runC(q)
	if err != nil {
		return insertFailure(err)
	}

	return nil
}

func UpdateComputer(oid addie.Id, old addie.Computer, c addie.Computer) (int, error) {

	key, err := UpdateNetworkHost(oid, old.NetHost, c.NetHost)
	if err != nil {
		return -1, updateFailure(err)
	}

	q := fmt.Sprintf("SELECT position_id FROM computers WHERE id = %d", key)
	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		return key, selectFailure(err)
	}
	if !rows.Next() {
		return key, emptyReadFailure()
	}
	var pos_key int
	err = rows.Scan(&pos_key)
	if err != nil {
		return key, scanFailure(err)
	}

	_, err = UpdatePosition(pos_key, c.Position)
	if err != nil {
		return key, updateFailure(err)
	}

	q = fmt.Sprintf(
		"UPDATE computers SET os = '%s', start_script = '%s' WHERE id = %d",
		c.OS, c.Start_script, key)

	err = runC(q)
	if err != nil {
		return key, updateFailure(err)
	}

	return key, nil

}

func ReadComputerByKey(id_key int) (*addie.Computer, error) {

	id, err := ReadId(id_key)
	if err != nil {
		return nil, readFailure(err)
	}

	q := fmt.Sprintf(
		"SELECT os, start_script, position_id FROM computers WHERE id = %d", id_key)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		return nil, selectFailure(err)
	}

	if !rows.Next() {
		return nil, emptyReadFailure()
	}
	var os, start_script string
	var pos_key int
	err = rows.Scan(&os, &start_script, &pos_key)
	if err != nil {
		return nil, scanFailure(err)
	}
	rows.Close()

	pos, err := ReadPosition(pos_key)
	if err != nil {
		return nil, readFailure(err)
	}

	ifs, err := ReadHostInterfaces(id_key)
	if err != nil {
		return nil, readFailure(err)
	}

	c := addie.Computer{}
	c.Id = *id
	c.Interfaces = *ifs
	c.OS = os
	c.Start_script = start_script
	c.Position = *pos

	return &c, nil

}

func ReadComputer(id addie.Id) (*addie.Computer, error) {

	id_key, err := ReadIdKey(id)
	if err != nil {
		return nil, readFailure(err)
	}

	return ReadComputerByKey(id_key)

}

// Routers ---------------------------------------------------------------------------

func CreateRouter(r addie.Router) error {

	//create nethost
	host_key, err := CreateNetworkHost(r.NetHost)
	if err != nil {
		return createFailure(err)
	}

	//position insert
	pos_key, err := CreatePosition(r.Position)
	if err != nil {
		return createFailure(err)
	}

	//packet conductor insert
	pkt_key, err := CreatePacketConductor(r.PacketConductor)
	if err != nil {
		return createFailure(err)
	}

	//router insert
	q := fmt.Sprintf(
		"INSERT INTO routers (id, packet_conductor_id, position_id) "+
			"values (%d, %d, %d)",
		host_key, pkt_key, pos_key)

	err = runC(q)
	if err != nil {
		return insertFailure(err)
	}

	return nil
}

func ReadRouterByKey(key int) (*addie.Router, error) {

	id, err := ReadId(key)
	if err != nil {
		return nil, readFailure(err)
	}

	q := fmt.Sprintf(
		"SELECT packet_conductor_id, position_id FROM routers WHERE id = %d", key)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		return nil, selectFailure(err)
	}
	if !rows.Next() {
		return nil, emptyReadFailure()
	}

	var pkt_key, pos_key int
	err = rows.Scan(&pkt_key, &pos_key)
	if err != nil {
		return nil, scanFailure(err)
	}
	rows.Close()

	pos, err := ReadPosition(pos_key)
	if err != nil {
		return nil, readFailure(err)
	}

	pkt, err := ReadPacketConductor(pkt_key)
	if err != nil {
		return nil, readFailure(err)
	}

	ifs, err := ReadHostInterfaces(key)
	if err != nil {
		return nil, readFailure(err)
	}

	rtr := addie.Router{}
	rtr.Id = *id
	rtr.Interfaces = *ifs
	rtr.PacketConductor = *pkt
	rtr.Position = *pos

	return &rtr, nil

}

func ReadRouter(id addie.Id) (*addie.Router, error) {

	key, err := ReadIdKey(id)
	if err != nil {
		return nil, readFailure(err)
	}

	return ReadRouterByKey(key)
}

func UpdateRouter(oid addie.Id, old addie.Router, r addie.Router) (int, error) {

	key, err := UpdateNetworkHost(oid, old.NetHost, r.NetHost)
	if err != nil {
		return -1, updateFailure(err)
	}

	q := fmt.Sprintf("SELECT position_id, packet_conductor_id FROM routers "+
		"WHERE id = %d", key)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		return key, selectFailure(err)
	}
	if !rows.Next() {
		return key, emptyReadFailure()
	}
	var pos_key, pkt_key int
	err = rows.Scan(&pos_key, &pkt_key)
	if err != nil {
		return key, scanFailure(err)
	}

	_, err = UpdatePosition(pos_key, r.Position)
	if err != nil {
		return key, updateFailure(err)
	}

	_, err = UpdatePacketConductor(pkt_key, r.PacketConductor)
	if err != nil {
		return key, updateFailure(err)
	}

	return key, nil
}

func ReadSystemRouters(key int) ([]addie.Router, error) {

	var result []addie.Router

	q := fmt.Sprintf(
		"SELECT routers.id FROM routers "+
			"INNER JOIN ids on routers.id = ids.id "+
			"WHERE ids.sys_id = %d", key)

	rows, err := runQ(q)
	defer safeClose(rows)

	if err != nil {
		return nil, selectFailure(err)
	}

	for rows.Next() {
		var rtr_key int
		err := rows.Scan(&rtr_key)
		if err != nil {
			return nil, scanFailure(err)
		}

		rtr, err := ReadRouterByKey(rtr_key)
		if err != nil {
			return nil, readFailure(err)
		}
		result = append(result, *rtr)
	}

	return result, nil

}

// Switches --------------------------------------------------------------------------

func CreateSwitch(s addie.Switch) error {

	//id insert
	id_key, err := CreateNetworkHost(s.NetHost)
	if err != nil {
		return createFailure(err)
	}

	//packet conductor insert
	pkt_key, err := CreatePacketConductor(s.PacketConductor)
	if err != nil {
		return createFailure(err)
	}

	//position insert
	pos_key, err := CreatePosition(s.Position)
	if err != nil {
		return createFailure(err)
	}

	//switch insert
	q := fmt.Sprintf(
		"INSERT INTO switches (id, packet_conductor_id, position_id)  "+
			"values (%d, %d, %d)",
		id_key, pkt_key, pos_key)

	err = runC(q)
	if err != nil {
		return insertFailure(err)
	}

	return nil
}

func ReadSwitchByKey(key int) (*addie.Switch, error) {

	id, err := ReadId(key)
	if err != nil {
		return nil, readFailure(err)
	}

	q := fmt.Sprintf(
		"SELECT packet_conductor_id, position_id FROM switches WHERE id = %d", key)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		return nil, selectFailure(err)
	}
	if !rows.Next() {
		return nil, emptyReadFailure()
	}

	var pkt_key, pos_key int
	err = rows.Scan(&pkt_key, &pos_key)
	if err != nil {
		return nil, scanFailure(err)
	}
	rows.Close()

	pos, err := ReadPosition(pos_key)
	if err != nil {
		return nil, readFailure(err)
	}

	pkt, err := ReadPacketConductor(pkt_key)
	if err != nil {
		return nil, readFailure(err)
	}

	ifs, err := ReadHostInterfaces(key)
	if err != nil {
		return nil, readFailure(err)
	}

	sw := addie.Switch{}
	sw.Id = *id
	sw.Interfaces = *ifs
	sw.PacketConductor = *pkt
	sw.Position = *pos

	return &sw, nil

}

func ReadSwitch(id addie.Id) (*addie.Switch, error) {

	key, err := ReadIdKey(id)
	if err != nil {
		return nil, readFailure(err)
	}

	return ReadSwitchByKey(key)
}

func UpdateSwitch(oid addie.Id, old addie.Switch, s addie.Switch) (int, error) {

	key, err := UpdateNetworkHost(oid, old.NetHost, s.NetHost)
	if err != nil {
		return -1, updateFailure(err)
	}

	q := fmt.Sprintf("SELECT packet_conductor_id, position_id FROM switches "+
		"WHERE id = %d", key)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		return key, selectFailure(err)
	}
	if !rows.Next() {
		return key, emptyReadFailure()
	}
	var pos_key, pkt_key int
	err = rows.Scan(&pkt_key, &pos_key)
	if err != nil {
		return key, scanFailure(err)
	}

	_, err = UpdatePosition(pos_key, s.Position)
	if err != nil {
		return key, updateFailure(err)
	}

	_, err = UpdatePacketConductor(pkt_key, s.PacketConductor)
	if err != nil {
		return key, updateFailure(err)
	}

	return key, nil

}

func ReadSystemSwitches(key int) ([]addie.Switch, error) {

	var result []addie.Switch

	q := fmt.Sprintf(
		"SELECT switches.id FROM switches "+
			"INNER JOIN ids on switches.id = ids.id "+
			"WHERE ids.sys_id = %d", key)

	rows, err := runQ(q)
	defer safeClose(rows)

	if err != nil {
		return nil, selectFailure(err)
	}

	for rows.Next() {
		var sw_key int
		err := rows.Scan(&sw_key)
		if err != nil {
			return nil, scanFailure(err)
		}

		sw, err := ReadSwitchByKey(sw_key)
		if err != nil {
			return nil, readFailure(err)
		}
		result = append(result, *sw)
	}

	return result, nil

}

// Links -----------------------------------------------------------------------------

func CreateLink(l addie.Link) error {

	//id insert
	id_key, err := CreateId(l.Id)
	if err != nil {
		return createFailure(err)
	}

	//packet conductor insert
	pkt_key, err := CreatePacketConductor(l.PacketConductor)
	if err != nil {
		return createFailure(err)
	}

	//endpoint0
	ep0_key, err := ReadIdKey(l.Endpoints[0].Id)
	if err != nil {
		return readFailure(err)
	}
	if0_key, err := ReadInterfaceKey(ep0_key, l.Endpoints[0].IfName)
	if err != nil {
		return readFailure(err)
	}

	//endpoint1
	ep1_key, err := ReadIdKey(l.Endpoints[1].Id)
	if err != nil {
		return readFailure(err)
	}
	if1_key, err := ReadInterfaceKey(ep1_key, l.Endpoints[1].IfName)
	if err != nil {
		return readFailure(err)
	}

	q := fmt.Sprintf(
		"INSERT INTO links "+
			"(id, packet_conductor_id, "+
			"endpoint_a_id, interface_a_id, "+
			"endpoint_b_id, interface_b_id) "+
			"VALUES (%d, %d, %d, %d, %d, %d)",
		id_key, pkt_key, ep0_key, if0_key, ep1_key, if1_key)

	err = runC(q)
	if err != nil {
		return insertFailure(err)
	}

	return nil

}

func ReadLinkByKey(key int) (*addie.Link, error) {

	id, err := ReadId(key)
	if err != nil {
		return nil, readFailure(err)
	}

	q := fmt.Sprintf(
		"SELECT packet_conductor_id, "+
			"endpoint_a_id, interface_a_id, "+
			"endpoint_b_id, interface_b_id "+
			"FROM links WHERE id = %d",
		key)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		return nil, selectFailure(err)
	}
	if !rows.Next() {
		return nil, emptyReadFailure()
	}

	var pkt_key, ep0_key, if0_key, ep1_key, if1_key int
	err = rows.Scan(&pkt_key, &ep0_key, &if0_key, &ep1_key, &if1_key)
	if err != nil {
		return nil, scanFailure(err)
	}
	rows.Close()

	pkt, err := ReadPacketConductor(pkt_key)
	if err != nil {
		return nil, readFailure(err)
	}

	ep0, err := ReadId(ep0_key)
	if err != nil {
		return nil, readFailure(err)
	}

	ep1, err := ReadId(ep1_key)
	if err != nil {
		return nil, readFailure(err)
	}

	if0, err := ReadInterface(if0_key)
	if err != nil {
		return nil, readFailure(err)
	}

	if1, err := ReadInterface(if1_key)
	if err != nil {
		return nil, readFailure(err)
	}

	lnk := addie.Link{}
	lnk.Id = *id
	lnk.PacketConductor = *pkt
	lnk.Endpoints[0] = addie.NetIfRef{*ep0, if0.Name}
	lnk.Endpoints[1] = addie.NetIfRef{*ep1, if1.Name}

	return &lnk, nil

}

func ReadLink(id addie.Id) (*addie.Link, error) {

	key, err := ReadIdKey(id)
	if err != nil {
		return nil, readFailure(err)
	}

	return ReadLinkByKey(key)
}

func UpdateLink(oid addie.Id, l addie.Link) (int, error) {

	key, err := UpdateId(oid, l.Id)
	if err != nil {
		return -1, updateFailure(err)
	}

	q := fmt.Sprintf("SELECT packet_conductor_id FROM links WHERE id = %d", key)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		return key, selectFailure(err)
	}
	if !rows.Next() {
		return key, emptyReadFailure()
	}

	var pkt_key int
	err = rows.Scan(&pkt_key)
	if err != nil {
		return key, scanFailure(err)
	}

	//packet conductor
	_, err = UpdatePacketConductor(pkt_key, l.PacketConductor)
	if err != nil {
		return key, updateFailure(err)
	}

	//endpoints
	e0, err := ReadIdKey(l.Endpoints[0].Id)
	if err != nil {
		return key, readFailure(err)
	}
	e1, err := ReadIdKey(l.Endpoints[1].Id)
	if err != nil {
		return key, readFailure(err)
	}

	//interfaces
	i0, err := ReadInterfaceKey(e0, l.Endpoints[0].IfName)
	if err != nil {
		return key, readFailure(err)
	}
	i1, err := ReadInterfaceKey(e1, l.Endpoints[1].IfName)
	if err != nil {
		return key, readFailure(err)
	}

	q = fmt.Sprintf(
		"UPDATE links SET "+
			"endpoint_a_id = %d, interface_a_id = %d, "+
			"endpoint_b_id = %d, interface_b_id = %d "+
			"WHERE id = %d", e0, i0, e1, i1, key)
	err = runC(q)
	if err != nil {
		return key, updateFailure(err)
	}

	return key, nil
}

func ReadSystemLinks(key int) ([]addie.Link, error) {

	var result []addie.Link

	q := fmt.Sprintf(
		"SELECT links.id FROM links "+
			"INNER JOIN ids on links.id = ids.id "+
			"WHERE ids.sys_id = %d", key)

	rows, err := runQ(q)
	defer safeClose(rows)

	if err != nil {
		return nil, selectFailure(err)
	}

	for rows.Next() {
		var lnk_key int
		err := rows.Scan(&lnk_key)
		if err != nil {
			return nil, scanFailure(err)
		}

		lnk, err := ReadLinkByKey(lnk_key)
		if err != nil {
			return nil, readFailure(err)
		}
		result = append(result, *lnk)
	}

	return result, nil

}
