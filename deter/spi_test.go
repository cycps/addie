package deter

import (
	"github.com/cycps/addie/db"
	"testing"
)

func TestSPI(t *testing.T) {

	dsg, err := db.ReadDesign("phobos", "murphy")
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Failed to read design")
	}

}
