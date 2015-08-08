package deter

import (
	"encoding/xml"
	"github.com/cycps/addie/db"
	"testing"
)

func TestSPI(t *testing.T) {

	dsg, err := db.ReadDesign("chinook", "murphy")
	if err != nil {
		t.Log(err)
		t.Fatal("Failed to read design")
	}

	t.Log(dsg)

	topo := designTopDL(dsg)
	topdl, err := xml.MarshalIndent(topo, "  ", "  ")
	if err != nil {
		t.Log(err)
		t.Error("failed to serialize topology to topdl xml")
	}
	t.Log(string(topdl))

}
