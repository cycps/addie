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
	rows, err := db.Query(q)
	if err != nil {
		log.Println(err)
		return nil, errors.New("Failed to run query")
	}
	return rows, nil
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

func InsertComputer(c addie.Computer) error {

	//TODO you are here

	sys_id, err := SysKey(c.Sys) //TODO
	if err != nil {
		log.Println(err)
		return errors.New("retrieving system " + c.Sys + " failed")
	}
	design_id, err := DesignKey(c.Design) //TODO
	if err != nil {
		log.Println(err)
		return errors.New("retrieving design " + c.Design + " failed")
	}

	q := fmt.Sprintf(
		"INSERT INTO ids (name, sys_id, design_id) VALUES ('%s', %d, %d)",
		c.Name, sys_id, design_id)

	return nil
}
