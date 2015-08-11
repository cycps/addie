/*
This file contians the code for generating a cypress simulation description
from a design in the experiment database
*/
package sim

import (
	"github.com/cycps/addie"
)

/*The GenerateSourceFromDB function generates Cypress simulation source for a
design given its name. This function reads the design from the database based
on the provided name. If you already have the design in memory use the
GenerateSource function.
*/
func GenerateSourceFromDB(name string) string {

	src := ""

	return src
}

/*The GenerateSource function generates Cypress simulation source for a design
given a pointer to an in memory design object.
*/
func GenerateSource(dsg *addie.Design) string {

	src := ""

	return src

}
