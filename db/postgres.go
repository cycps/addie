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
)

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
	return nil
}

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

func GetDesigns() (map[string]struct{}, error) {
	m := make(map[string]struct{})

	rows, err := runQ("SELECT (name) FROM designs")
	defer safeClose(rows)
	if err != nil {
		return nil, err
	}

	var e struct{}

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			log.Println(err)
			return nil, errors.New("could not scan the row")
		}

		m[name] = e
	}

	return m, nil
}

func InsertDesign(name string) error {

	q := fmt.Sprintf("INSERT INTO designs (name) VALUES ('%s')", name)
	err := runC(q)
	if err != nil {
		return err
	}

	return nil
}

func TrashDesign(name string) error {

	q := fmt.Sprintf("DELETE FROM designs WHERE name = '%s'", name)
	err := runC(q)
	if err != nil {
		return err
	}

	return nil
}

func InsertSystem(design string, name string) (int, error) {

	design_key, err := DesignKey(design)
	if err != nil {
		log.Println(err)
		return -1, fmt.Errorf("[InsertSystem] design '%s' does not exist", design)
	}

	q := fmt.Sprintf(
		"INSERT INTO systems (design_id, name) VALUES (%d, '%s') RETURNING id",
		design_key, name)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		log.Println(err)
		return -1, fmt.Errorf("[InsertSystem] insertion query failed")
	}
	if !rows.Next() {
		return -1, fmt.Errorf("[InsertSystem] insertion readback failed")
	}
	var id_key int
	err = rows.Scan(&id_key)
	if err != nil {
		log.Println(err)
		return -1, fmt.Errorf("[InsertSystem] id readback scan failed")
	}

	return id_key, nil

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

func SysKey(design string, name string) (int, error) {

	design_key, err := DesignKey(design)
	if err != nil {
		log.Println(err)
		return -1, fmt.Errorf("[SysKey] design '%s' does not exist", design)
	}

	q := fmt.Sprintf("SELECT id FROM systems WHERE name = '%s' AND design_id = %d",
		name, design_key)

	sys_key, err := getKey(q)
	if err != nil {
		log.Println(err)
		return -1, fmt.Errorf("could not get key for system '%s'", name)
	}
	return sys_key, nil
}

func IdKey(id addie.Id) (int, error) {

	sys_key, err := SysKey(id.Design, id.Sys)
	if err != nil {
		log.Println(err)
		return -1, fmt.Errorf(
			"[IdKey] (design, sys) combo ('%s', '%s') does not exist",
			id.Design, id.Sys)
	}

	q := fmt.Sprintf("SELECT id FROM ids WHERE name = '%s' AND sys_id = '%d'",
		id.Name, sys_key)

	id_key, err := getKey(q)
	if err != nil {
		log.Println(err)
		return -1, fmt.Errorf("could not get id key for %v", id)
	}
	return id_key, nil
}

func InterfaceKey(host_id int, ifname string) (int, error) {

	q := fmt.Sprintf(
		"SELECT id FROM interfaces WHERE host_id = %d AND name = '%s'",
		host_id, ifname)

	key, err := getKey(q)
	if err != nil {
		log.Println(err)
		return -1, fmt.Errorf(
			"could not get key for interface hostid=%d, ifname=%s",
			host_id, ifname)
	}

	return key, nil

}

func DesignKey(name string) (int, error) {

	q := fmt.Sprintf("SELECT id FROM designs WHERE name = '%s'", name)
	key, err := getKey(q)
	if err != nil {
		log.Println(err)
		return -1, fmt.Errorf("could not get key for design '%s'", name)
	}
	return key, nil
}

func InsertId(id addie.Id) (int, error) {
	sys_id, err := SysKey(id.Design, id.Sys)
	if err != nil {
		log.Println(err)
		return -1, fmt.Errorf("r[InsertId] retrieving system '%s' failed", id.Sys)
	}

	q := fmt.Sprintf(
		"INSERT INTO ids (name, sys_id) VALUES ('%s', %d) RETURNING id",
		id.Name, sys_id)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		log.Println(err)
		return -1, fmt.Errorf("[InsertId] id insert failed")
	}

	if !rows.Next() {
		return -1, fmt.Errorf("[InsertId] id readback failed")
	}
	var id_key int
	err = rows.Scan(&id_key)
	if err != nil {
		log.Println(err)
		return -1, fmt.Errorf("[InsertId] id readback scan failed")
	}

	return id_key, nil
}

func InsertNetworkHostByKey(id_key int) error {

	q := fmt.Sprintf(
		"INSERT INTO network_hosts (id) VALUES (%d)", id_key)

	err := runC(q)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertNetworkHostByKey] network_host insert failed")
	}

	return nil

}

func InsertPosition(p addie.Position) (int, error) {

	q := fmt.Sprintf(
		"INSERT INTO positions (x, y, z) VALUES (%f, %f, %f) RETURNING id",
		p.X, p.Y, p.Z)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		log.Println(err)
		return -1, fmt.Errorf("[InsertPosition] position insert failed")
	}
	if !rows.Next() {
		log.Println(err)
		return -1, fmt.Errorf("[InsertPosition] pg RETURNING cursor did not return anything")
	}

	var pos_key int
	err = rows.Scan(&pos_key)
	if err != nil {
		log.Println(err)
		return -1, fmt.Errorf("[InsertPosition] failed to read pg RETURNING row")
	}

	return pos_key, nil
}

func InsertPacketConductor(p addie.PacketConductor) (int, error) {

	q := fmt.Sprintf(
		"INSERT INTO packet_conductors (capacity, latency) VALUES (%d, %d) "+
			"RETURNING id",
		p.Capacity, p.Latency)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		log.Println(err)
		return -1, fmt.Errorf("[InsertPacketConductor] insert failed")
	}
	if !rows.Next() {
		log.Println(err)
		return -1, fmt.Errorf(
			"[InsertPacketConductor] pg RETURNING cursor returned nil")
	}

	var pkt_key int
	err = rows.Scan(&pkt_key)
	if err != nil {
		log.Println(err)
		return -1, fmt.Errorf("[InsertPacketConductor] faild to read pg RETURNING row")
	}

	return pkt_key, nil

}

func InsertInterface(host_id int, ifx addie.Interface) error {

	pkt_id, err := InsertPacketConductor(ifx.PacketConductor)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertInterface] unable to insert packet conductor")
	}

	q := fmt.Sprintf(
		"INSERT INTO interfaces (name, host_id, packet_conductor_id) "+
			"VALUES ('%s', %d, %d)",
		ifx.Name, host_id, pkt_id)

	err = runC(q)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertInterface] insert failed")
	}

	return nil
}

func GetInterface(id int) (*addie.Interface, error) {

	q := fmt.Sprintf(
		"SELECT name, host_id, packet_conductor_id FROM interfaces WHERE id = %d", id)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetInterface] error running query: %s", q)
	}
	if !rows.Next() {
		return nil, fmt.Errorf("[GetInterface] could not find id=%d", id)
	}
	var name string
	var host_key, pkt_key int
	err = rows.Scan(&name, &host_key, &pkt_key)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetInterface] error reading query result")
	}

	pkt, err := GetPacketConductor(pkt_key)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetInterface] error getting packet conductor")
	}

	ifx := addie.Interface{}
	ifx.Name = name
	ifx.PacketConductor = *pkt

	return &ifx, nil

}

func GetHostInterfaces(host_id int) (*map[string]addie.Interface, error) {

	q := fmt.Sprintf(
		"SELECT name, packet_conductor_id FROM interfaces WHERE host_id = %d", host_id)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetHostInterfaces] error running query")
	}
	result := make(map[string]addie.Interface)
	for rows.Next() {
		var name string
		var pkt_key int
		err = rows.Scan(&name, &pkt_key)
		if err != nil {
			log.Println(err)
			return nil, fmt.Errorf("[GetHostInterfaces] error reading query result")
		}

		pkt, err := GetPacketConductor(pkt_key)
		if err != nil {
			log.Println(err)
			return nil, fmt.Errorf("[GetHostInterfaces] error getting packet conductor")
		}

		ifx := addie.Interface{}
		ifx.Name = name
		ifx.PacketConductor = *pkt

		result[name] = ifx

	}

	return &result, nil
}

func InsertComputer(c addie.Computer) error {

	//id insert
	id_key, err := InsertId(c.Id)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertComputer] Id insert failed")
	}

	//network host insert
	err = InsertNetworkHostByKey(id_key)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertComputer] Network Host insert failed")
	}

	//position insert
	pos_key, err := InsertPosition(c.Position)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertComputer] Position insert failed")
	}

	//interfaces insert
	for _, ifx := range c.Interfaces {
		err = InsertInterface(id_key, ifx)
		if err != nil {
			log.Println(err)
			return fmt.Errorf("[InsertComputer] Failed to insert interface %s",
				ifx.Name)
		}
	}

	//computer insert
	q := fmt.Sprintf(
		"INSERT INTO computers (id, os, start_script, position_id) "+
			"values (%d, '%s', '%s', %d)",
		id_key, c.OS, c.Start_script, pos_key)

	err = runC(q)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Error inserting Computer '%s' into the DB", c.Name)
	}

	return nil
}

/*
UpdateID updates an id. If the system in the new id does not exist it is created.
Changing design is not supported through this interface
*/
func UpdateId(oid addie.Id, id addie.Id) error {

	if oid == id {
		return nil
	}
	if oid.Design != id.Design {
		return fmt.Errorf("[UpdateId] changing design though this interface not supported")
	}

	oid_key, err := IdKey(oid)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[UpdateId] bad oid")
	}

	if oid.Sys != id.Sys {
		sys_key, err := SysKey(id.Design, id.Sys)
		if err != nil {
			sys_key, err = InsertSystem(id.Design, id.Sys)
			if err != nil {
				log.Println(err)
				return fmt.Errorf("[UpdateId] fail to insert new system")
			}
		}
		q := fmt.Sprintf(
			"UPDATE ids SET sys_id = %d WHERE id = %d", sys_key, oid_key)

		err = runC(q)
		if err != nil {
			log.Println(err)
			return fmt.Errorf("[UpdateId] failed to set sys_id=%d", sys_key)
		}

		//err := SysRecycle() do this in background?
		if err != nil {
			log.Println(err)
			return fmt.Errorf("[UpdateId] an error occured during recycling")
		}
	}
	if oid.Name != id.Name {
		q := fmt.Sprintf(
			"UPDATE ids SET name = '%s' WHERE id = %d", id.Name, oid_key)
		err = runC(q)
		if err != nil {
			log.Println(err)
			return fmt.Errorf("[UpdateId] failed to set name=%d", id.Name)
		}
	}

	return nil
}

func SysRecycle() error {

	//TODO

	return nil

}

func UpdateComputer(oid addie.Id, c addie.Computer) error {

	if oid != c.Id {
		err := UpdateId(oid, c.Id)
		if err != nil {
			log.Println(err)
			return fmt.Errorf("[UpdateComputer] failed to update id")
		}
	}

	id_key, err := IdKey(c.Id)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[UpdateComputer] bad id")
	}

	q := fmt.Sprintf(
		"UPDATE computers SET os = '%s', start_script = '%s' WHERE id = %d",
		c.OS, c.Start_script, id_key)

	err = runC(q)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[UpdateComputer] failed to run computer update query")
	}

	//return fmt.Errorf("[UpdateComputer] not implememted")
	return nil

}

func InsertRouter(r addie.Router) error {

	//TODO: this is almost the exact same code as the beginning of ComputerInsert
	//make an interface to combine this shizzzz

	//id insert
	id_key, err := InsertId(r.Id)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertRouter] Id insert failed")
	}

	//network host insert
	err = InsertNetworkHostByKey(id_key)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertRouter] Network Host insert failed")
	}

	//position insert
	pos_key, err := InsertPosition(r.Position)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertRouter] Position insert failed")
	}

	//packet conductor insert
	pkt_key, err := InsertPacketConductor(r.PacketConductor)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertRouter] Packet conductor insert failed")
	}

	//router insert
	q := fmt.Sprintf(
		"INSERT INTO routers (id, packet_conductor_id, position_id) "+
			"values (%d, %d, %d)",
		id_key, pkt_key, pos_key)

	err = runC(q)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Error inserting Router '%s' intro the DB", r.Name)
	}

	return nil
}

func GetId(id int) (*addie.Id, error) {

	//fetch the id
	q := fmt.Sprintf(
		"SELECT name, sys_id FROM ids WHERE id = %d", id)
	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetId] id-query error: %s", q)
	}
	if !rows.Next() {
		return nil, fmt.Errorf("[GetId] identity with id=%d does not exist", id)
	}
	var id_name string
	var sys_key int
	err = rows.Scan(&id_name, &sys_key)
	if err != nil {
		return nil, fmt.Errorf("[GetId] failed to read id-query result")
	}

	//fetch the sys
	q = fmt.Sprintf(
		"SELECT name, design_id FROM systems WHERE id=%d", sys_key)
	rows, err = runQ(q)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetId] sys-query error: %s", q)
	}
	if !rows.Next() {
		return nil, fmt.Errorf("[GetId] sys with id=%d does not exist", id)
	}
	var sys_name string
	var dsg_key int
	err = rows.Scan(&sys_name, &dsg_key)
	if err != nil {
		return nil, fmt.Errorf("[GetId] failed to read sys-query result")
	}

	//fetch the design
	q = fmt.Sprintf(
		"SELECT name FROM designs WHERE id=%d", dsg_key)
	rows, err = runQ(q)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetId] design-query error: %s", q)
	}
	if !rows.Next() {
		return nil, fmt.Errorf("[GetId] design with id=%d does not exist", id)
	}
	var dsg_name string
	err = rows.Scan(&dsg_name)
	if err != nil {
		return nil, fmt.Errorf("[GetId] failed to read design-query result")
	}

	return &addie.Id{id_name, sys_name, dsg_name}, nil

}

func GetPosition(id int) (*addie.Position, error) {

	q := fmt.Sprintf(
		"SELECT x, y, z FROM positions WHERE id = %d", id)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetPosition] query error: %s", q)
	}

	if !rows.Next() {
		return nil, fmt.Errorf("[GetPosition] position with id=%d does not exist", id)
	}

	var x, y, z float32
	err = rows.Scan(&x, &y, &z)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetPosition] error reading result row")
	}

	return &addie.Position{x, y, z}, nil

}

func GetPacketConductor(id int) (*addie.PacketConductor, error) {

	q := fmt.Sprintf(
		"SELECT capacity, latency FROM packet_conductors WHERE id = %d", id)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetPacketConductor] query error: %s", q)
	}
	if !rows.Next() {
		return nil, fmt.Errorf("[GetPacketConductor] id=%d does not exist", id)
	}

	var capacity, latency int
	err = rows.Scan(&capacity, &latency)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetPacketConductor] error reading result row")
	}

	return &addie.PacketConductor{capacity, latency}, nil

}

func GetComputerByKey(id_key int) (*addie.Computer, error) {

	id, err := GetId(id_key)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetComputer] bad id key %d", id_key)
	}

	q := fmt.Sprintf(
		"SELECT os, start_script, position_id FROM computers WHERE id = %d", id_key)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetComputer] failed to run query: %s", q)
	}

	if !rows.Next() {
		return nil, fmt.Errorf("Failed find a computer with id %v", id)
	}
	var os, start_script string
	var pos_key int
	err = rows.Scan(&os, &start_script, &pos_key)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetComputer] failed to read row result")
	}

	pos, err := GetPosition(pos_key)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetComputer] failed to retrieve computer position")
	}

	ifs, err := GetHostInterfaces(id_key)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetComputer] failed to get computer interfaces")
	}

	c := addie.Computer{}
	c.Id = *id
	c.Interfaces = *ifs
	c.OS = os
	c.Start_script = start_script
	c.Position = *pos

	return &c, nil

}

func GetComputer(id addie.Id) (*addie.Computer, error) {

	id_key, err := IdKey(id)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf(
			"[GetComputer] unable to retrieve design or system for the id %v", id)
	}

	return GetComputerByKey(id_key)

}

func GetRouter(id addie.Id) (*addie.Router, error) {

	id_key, err := IdKey(id)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetRouter] get id %v failed", id)
	}

	q := fmt.Sprintf(
		"SELECT packet_conductor_id, position_id FROM routers WHERE id = %d", id_key)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetRouter] failed to run query: %s", q)
	}
	if !rows.Next() {
		return nil, fmt.Errorf("[GetRouter] Failed to find router with id %v", id)
	}

	var pkt_key, pos_key int
	err = rows.Scan(&pkt_key, &pos_key)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetRouter] failed to read row result")
	}

	pos, err := GetPosition(pos_key)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetRouter] failed to get position")
	}

	pkt, err := GetPacketConductor(pkt_key)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetRouter] failed to get packet conductor")
	}

	rtr := addie.Router{}
	rtr.Id = id
	rtr.Interfaces = make(map[string]addie.Interface) //todo
	rtr.PacketConductor = *pkt
	rtr.Position = *pos

	return &rtr, nil

}

func InsertSwitch(s addie.Switch) error {

	//id insert
	id_key, err := InsertId(s.Id)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertSwitch] Id insert failed")
	}

	//packet conductor insert
	pkt_key, err := InsertPacketConductor(s.PacketConductor)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertSwitch] Packet Conductor insert failed")
	}

	//position insert
	pos_key, err := InsertPosition(s.Position)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertRouter] Position insert failed")
	}

	//switch insert
	q := fmt.Sprintf(
		"INSERT INTO switches (id, packet_conductor_id, position_id)  "+
			"values (%d, %d, %d)",
		id_key, pkt_key, pos_key)

	err = runC(q)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Error inserting Switch '%s' intro the DB", s.Name)
	}

	return nil
}

func GetSwitch(id addie.Id) (*addie.Switch, error) {

	id_key, err := IdKey(id)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetSwitch] get id %v failed", id)
	}

	q := fmt.Sprintf(
		"SELECT packet_conductor_id, position_id FROM switches WHERE id = %d", id_key)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetSwitch] failed to run query: %s", q)
	}
	if !rows.Next() {
		return nil, fmt.Errorf("[GetSwitch] Failed to find switch with id %v", id)
	}

	var pkt_key, pos_key int
	err = rows.Scan(&pkt_key, &pos_key)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetSwitch] failed to read row result")
	}

	pos, err := GetPosition(pos_key)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetSwitch] failed to get position")
	}

	pkt, err := GetPacketConductor(pkt_key)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetSwitch] failed to get packet conductor")
	}

	sw := addie.Switch{}
	sw.Id = id
	sw.PacketConductor = *pkt
	sw.Position = *pos

	return &sw, nil

}

func InsertLink(l addie.Link) error {

	//id insert
	id_key, err := InsertId(l.Id)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertLink] Id insert failed")
	}

	//packet conductor insert
	pkt_key, err := InsertPacketConductor(l.PacketConductor)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertLink] Packet Conductor insert failed")
	}

	//endpoint0
	ep0_key, err := IdKey(l.Endpoints[0].Id)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertLink] bad endpoint[0]")
	}
	if0_key, err := InterfaceKey(ep0_key, l.Endpoints[0].IfName)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertLink] bad endpoint[0] interface")
	}

	//endpoint1
	ep1_key, err := IdKey(l.Endpoints[1].Id)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertLink] bad endpoint[1]")
	}
	if1_key, err := InterfaceKey(ep0_key, l.Endpoints[0].IfName)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertLink] bad endpoint[1] interface")
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
		log.Println(err)
		return fmt.Errorf("Error inserting link '%s' into the DB", l.Name)
	}

	return nil

}

func GetLink(id addie.Id) (*addie.Link, error) {

	id_key, err := IdKey(id)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetLink] get id %v failed", id)
	}

	q := fmt.Sprintf(
		"SELECT packet_conductor_id, "+
			"endpoint_a_id, interface_a_id, "+
			"endpoint_b_id, interface_b_id "+
			"FROM links WHERE id = %d",
		id_key)

	rows, err := runQ(q)
	defer safeClose(rows)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetLink] failed to run query: %s", q)
	}
	if !rows.Next() {
		return nil, fmt.Errorf("[GetLink] failed to find link with id %v", id)
	}

	var pkt_key, ep0_key, if0_key, ep1_key, if1_key int
	err = rows.Scan(&pkt_key, &ep0_key, &if0_key, &ep1_key, &if1_key)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetLink] failed to read row result")
	}

	pkt, err := GetPacketConductor(pkt_key)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetLink] failed to get packet conductor")
	}

	ep0, err := GetId(ep0_key)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetLink] failed to get endpoint[1]")
	}

	ep1, err := GetId(ep1_key)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetLink] failed to get endpoint[0]")
	}

	if0, err := GetInterface(if0_key)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetLink] failed to get endpoint[0] interface")
	}

	if1, err := GetInterface(if1_key)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetLink] failed to get endpoint[1] interface")
	}

	lnk := addie.Link{}
	lnk.Id = id
	lnk.PacketConductor = *pkt
	lnk.Endpoints[0] = addie.NetIfRef{*ep0, if0.Name}
	lnk.Endpoints[1] = addie.NetIfRef{*ep1, if1.Name}

	return &lnk, nil

}
