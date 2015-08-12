/*
This file contians the code for generating a cypress simulation description
from a design in the experiment database
*/
package sim

import (
	"fmt"
	"github.com/cycps/addie"
	"github.com/cycps/addie/db"
	"log"
)

/*The GenerateSourceFromDB function generates Cypress simulation source for a
design given its name. This function reads the design from the database based
on the provided name. If you already have the design in memory use the
GenerateSource function.
*/
func GenerateSourceFromDB(designName, user string) (string, error) {

	dsg, err := db.ReadDesign(designName, user)
	if err != nil {
		log.Println(err)
		return "", fmt.Errorf("Failed to read design")
	}

	models, err := db.ReadUserModels(user)
	if err != nil {
		log.Println(err)
		return "", fmt.Errorf("failed to read user models")
	}

	return GenerateSource(dsg, models), nil

}

/*The GenerateSource function generates Cypress simulation source for a design
given a pointer to an in memory design object.
*/
func GenerateSource(dsg *addie.Design, models []addie.Model) string {

	src := ""

	for i, _ := range models {
		src += modelSrc(&models[i])
	}

	return src

}

func modelSrc(m *addie.Model) string {

	src := "Object " + m.Name + "\n" +
		m.Equations

	return src

}
