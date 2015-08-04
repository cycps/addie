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

/*
func runC(q string) (sql.Result, error) {
	err := dbPing()
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to execute command -- error communicating with DB")
	}
	result, err := db.Exec(q)
	if err != nil {
		log.Println(err)
		return nil, errors.New("Failed to run execute command")
	}
	return result, nil
}
*/

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
	if err != nil {
		return nil, err
	}

	defer rows.Close()
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
	_, err := runQ(q)
	if err != nil {
		return err
	}

	return nil
}

func TrashDesign(name string) error {

	q := fmt.Sprintf("DELETE FROM designs WHERE name = '%s'", name)
	_, err := runQ(q)
	if err != nil {
		return err
	}

	return nil
}

func InsertSystem(design string, name string) error {

	design_key, err := DesignKey(design)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertSystem] design '%s' does not exist", design)
	}

	q := fmt.Sprintf("INSERT INTO systems (design_id, name) VALUES (%d, '%s')",
		design_key, name)

	_, err = runQ(q)
	if err != nil {
		return err
	}

	return nil

}

func getKey(q string) (int, error) {

	rows, err := runQ(q)
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

func SysKey(design string, name string) (int, int, error) {

	design_key, err := DesignKey(design)
	if err != nil {
		log.Println(err)
		return -1, -1, fmt.Errorf("[SysKey] design '%s' does not exist", design)
	}

	q := fmt.Sprintf("SELECT id FROM systems WHERE name = '%s' AND design_id = %d",
		name, design_key)

	sys_key, err := getKey(q)
	if err != nil {
		log.Println(err)
		return -1, -1, fmt.Errorf("could not get key for system '%s'", name)
	}
	return design_key, sys_key, nil
}

func IdKey(id addie.Id) (int, int, int, error) {

	design_key, sys_key, err := SysKey(id.Design, id.Sys)
	if err != nil {
		log.Println(err)
		return -1, -1, -1, fmt.Errorf(
			"[IdKey] (design, sys) combo ('%s', '%s') does not exist",
			id.Design, id.Sys)
	}

	q := fmt.Sprintf("SELECT id FROM ids WHERE name = '%s' AND sys_id = '%d'",
		id.Name, sys_key)

	id_key, err := getKey(q)
	if err != nil {
		log.Println(err)
		return -1, -1, -1, fmt.Errorf("could not get id key for %v", id)
	}
	return design_key, sys_key, id_key, nil
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

func InsertComputer(c addie.Computer) error {

	_, sys_id, err := SysKey(c.Design, c.Sys)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("retrieving system '%s' failed", c.Sys)
	}

	//id insert
	q := fmt.Sprintf(
		"INSERT INTO ids (name, sys_id) VALUES ('%s', %d)",
		c.Name, sys_id)

	_, err = runQ(q)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertComputer] id insert failed")
	}

	_, _, id_key, err := IdKey(c.Id)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Computer insert failed -- could not retrieve key")
	}

	//network_host insert
	q = fmt.Sprintf(
		"INSERT INTO network_hosts (id) VALUES (%d)", id_key)

	_, err = runQ(q)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertComputer] network_host insert failed")
	}

	//pos insert
	q = fmt.Sprintf(
		"INSERT INTO positions (x, y, z) VALUES (%f, %f, %f) RETURNING id",
		c.Position.X, c.Position.Y, c.Position.Z)

	rows, err := runQ(q)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertComputer] position insert failed")
	}
	if !rows.Next() {
		log.Println(err)
		return fmt.Errorf("[InsertComputer] pg RETURNING cursor did not return anything")
	}

	var pos_key int
	err = rows.Scan(&pos_key)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("[InsertComputer] failed to read pg RETURNING row")
	}

	//computer insert
	q = fmt.Sprintf(
		"INSERT INTO computers (id, os, start_script, position_id) "+
			"values (%d, '%s', '%s', %d)",
		id_key, c.OS, c.Start_script, pos_key)

	_, err = runQ(q)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Error inserting Computer '%s' into the DB", c.Name)
	}

	return nil
}

func GetPosition(id int) (*addie.Position, error) {

	q := fmt.Sprintf(
		"SELECT x, y, z FROM positions WHERE id = %d", id)

	rows, err := runQ(q)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("[GetPosition] query error")
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

func GetComputer(id addie.Id) (*addie.Computer, error) {

	_, _, id_key, err := IdKey(id)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf(
			"[GetComputer] unable to retrieve design or system for the id %v", id)
	}

	q := fmt.Sprintf(
		"SELECT os, start_script, position_id FROM computers WHERE id = %d", id_key)

	rows, err := runQ(q)
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

	c := addie.Computer{}
	c.Id = id
	c.Interfaces = make(map[string]addie.Interface) //todo
	c.OS = os
	c.Start_script = start_script
	c.Position = *pos

	return &c, nil
}
