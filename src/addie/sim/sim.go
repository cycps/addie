/*
This file contians the code for generating a cypress simulation description
from a design in the experiment database

ASSUMPTION :!: We are assuming that the models and design have already been
							 semantically checked. Bad input here can be catastrophic.
*/
package sim

import (
	"addie"
	"addie/db"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
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

	src += designSrc(dsg)

	return src

}

func modelSrc(m *addie.Model) string {

	src := "Object " + m.Name + "(" + strings.TrimSuffix(m.Params, ",") + ")\n"

	eqtns := strings.Split(m.Equations, "\n")
	for _, e := range eqtns {
		src += "  " + e + "\n"
	}

	src += "\n"

	return src

}

func designSrc(d *addie.Design) string {

	src := "Simulation " + d.Name + "\n"

	var plinks []*addie.Plink

	for _, v := range d.Elements {

		switch v.(type) {

		case addie.Phyo:
			p := v.(addie.Phyo)
			src += phyoSrc(&p)

		case addie.Sax:
			s := v.(addie.Sax)
			src += saxSrc(&s)

		case addie.Plink:
			p := v.(addie.Plink)
			plinks = append(plinks, &p)

		}

	}

	src += "\n"

	for _, p := range plinks {
		src += plinkSrc(p, d)
	}

	return src

}

func phyoSrc(p *addie.Phyo) string {

	src := "  " + p.Model + " " + p.Name + "("
	src += strings.Replace(strings.TrimSuffix(p.Args, ","), "=", ":", -1)
	_init := strings.Replace(p.Init, " ", "", -1)
	if len(_init) > 0 {
		src += "," + strings.Replace(strings.TrimSuffix(_init, ","), "=", "|", -1)
	}
	src += ")"

	src += "\n"
	return src

}

func saxSrc(sax *addie.Sax) string {

	src := ""

	sensors := strings.Split(sax.Sense, ";")
	actuators := strings.Split(sax.Actuate, ";")

	for _, s := range sensors {
		src += sensorSrc(sax, s)
	}

	for _, a := range actuators {
		src += actuatorSrc(sax, a)
	}

	return src

}

func sensorSrc(sax *addie.Sax, s string) string {

	re, _ := regexp.Compile("([a-zA-Z_][a-zA-Z0-9_]*)\\(([0-9]*)\\)")
	m := re.FindAllStringSubmatch(s, -1)

	src := "  Sensor " + sax.Name + "_S_" + m[0][1] + "(Rate:" + m[0][2] +
		", Destination:localhost)"

	src += "\n"

	return src

}

func containsSensor(sax *addie.Sax, name string) bool {

	re, _ := regexp.Compile("([a-zA-Z_][a-zA-Z0-9_]*)\\(([0-9]*)\\)")

	sensors := strings.Split(sax.Sense, ";")

	for _, s := range sensors {
		m := re.FindAllStringSubmatch(s, -1)
		if m[0][1] == name {
			return true
		}
	}

	return false
}

func actuatorSrc(sax *addie.Sax, a string) string {

	re, _ := regexp.Compile(
		"([a-zA-Z_][a-zA-Z0-9_]*)\\(([0-9]+(?:\\.[0-9]+)?),([0-9]+(?:\\.[0-9]+)?)\\)")
	m := re.FindAllStringSubmatch(strings.Replace(a, " ", "", -1), -1)

	src := "  Actuator " + sax.Name + "_A_" + m[0][1] + "(" +
		"Min:-" + m[0][2] + ", " +
		"Max:" + m[0][2] + ", " +
		"DMin:-" + m[0][3] + ", " +
		"DMax:" + m[0][3] + ")"

	src += "\n"

	return src

}

func containsActuator(sax *addie.Sax, name string) bool {

	actuators := strings.Split(sax.Actuate, ";")

	re, _ := regexp.Compile(
		"([a-zA-Z_][a-zA-Z0-9_]*)\\(([0-9]+(?:\\.[0-9]+)?),([0-9]+(?:\\.[0-9]+)?)\\)")

	for _, a := range actuators {
		m := re.FindAllStringSubmatch(strings.Replace(a, " ", "", -1), -1)
		if m[0][1] == name {
			return true
		}
	}

	return false
}

func plinkSrc(plink *addie.Plink, d *addie.Design) string {

	src := ""

	aVars := strings.Split(strings.Replace(plink.Bindings[0], " ", "", -1), ",")
	bVars := strings.Split(strings.Replace(plink.Bindings[1], " ", "", -1), ",")

	ae := d.Elements[plink.Endpoints[0]]
	aName := ae.Identify().Name

	be := d.Elements[plink.Endpoints[1]]
	bName := be.Identify().Name

	for i, a := range aVars {

		b := bVars[i]

		if reflect.TypeOf(ae).Name() == "Sax" {
			s := ae.(addie.Sax)
			src += "  " + saxName(&s, a)
		} else {
			src += "  " + aName + "." + a
		}

		src += " ~ "

		if reflect.TypeOf(be).Name() == "Sax" {
			s := be.(addie.Sax)
			src += saxName(&s, b)
		} else {
			src += bName + "." + b
		}

		src += "\n"
	}

	return src

}

func saxName(sax *addie.Sax, name string) string {

	saxT := "?"
	saxV := "?"

	if containsSensor(sax, name) {
		saxT = "S"
		saxV = "y"
	} else if containsActuator(sax, name) {
		saxT = "A"
		saxV = "u"
	}

	src := sax.Name + "_" + saxT + "_" + name + "." + saxV

	return src

}
