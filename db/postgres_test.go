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

func TestOneCreateDestroy(t *testing.T) {

	err := InsertDesign("caprica")
	if err != nil {
		t.Error("failed to create caprica")
	}

	c := addie.Computer{}
	//Id
	c.Name = "c"
	c.Sys = "root"
	c.Design = "one"
	//NetHost
	c.Interfaces = make(map[string]addie.Interface)
	//Comptuer
	c.Position = addie.Position{0, 0, 0}
	c.OS = "Ubuntu-15.04"
	c.Start_script = "make_muffins.sh"

	err = InsertComputer(c)
	if err != nil {
		t.Error("failed to insert computer")
	}

	err = TrashDesign("caprica")
	if err != nil {
		t.Error("failed to trash caprica")
	}

}
