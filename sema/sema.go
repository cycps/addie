/*
The sema package performs semantic checking on Cypress models before they are compiled
by the PNetDL and TopDL compilers
*/
package sema

import (
	"fmt"
	"github.com/cycps/addie"
	"regexp"
	"strconv"
	"strings"
)

type Diagnostic struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

type Diagnostics struct {
	Elements []Diagnostic `json:"elements"`
}

func (ds *Diagnostics) Fatal() bool {
	for _, d := range ds.Elements {
		if d.Level == "error" {
			return true
		}
	}
	return false
}

func (ds *Diagnostics) Merge(x *Diagnostics) {
	ds.Elements = append(ds.Elements, x.Elements...)
}

func (ds *Diagnostics) ApplySource(s string) {

	for i, _ := range ds.Elements {
		ds.Elements[i].Message = strings.Replace(ds.Elements[i].Message,
			"$source", s, -1)
	}

}

func Check(dsg *addie.Design) Diagnostics {

	var ds Diagnostics

	/*
		ds.Elements = append(ds.Elements,
			Diagnostic{"info", "Do you know the muffin man?"})
	*/

	_ds := CheckPlinks(dsg)
	ds.Merge(&_ds)

	if !ds.Fatal() {
		ds.Elements = append(ds.Elements,
			Diagnostic{"success", "Design check succeeded"})
	}

	return ds

}

func CheckPlinks(dsg *addie.Design) Diagnostics {

	var ds Diagnostics

	for _, e := range dsg.Elements {
		switch e.(type) {
		case addie.Plink:
			p := e.(addie.Plink)
			_ds := CheckPlink(p, dsg)
			ds.Merge(&_ds)
		}
	}

	if ds.Fatal() {
		return ds
	}

	return ds

}

func CheckPlink(p addie.Plink, dsg *addie.Design) Diagnostics {

	var ds Diagnostics

	//check that endpoints exist
	e0, ok := dsg.Elements[p.Endpoints[0]]
	if !ok {
		ds.Elements = append(ds.Elements,
			Diagnostic{"error",
				fmt.Sprintf("[Plink][%v] references non-existant id [%v]", p.Id, p.Endpoints[0])})
	}
	e1, ok := dsg.Elements[p.Endpoints[1]]
	if !ok {
		ds.Elements = append(ds.Elements,
			Diagnostic{"error",
				fmt.Sprintf("[Plink][%v] references non-existant id [%v]", p.Id, p.Endpoints[1])})
	}
	if ds.Fatal() {
		return ds
	}

	_ds := CheckEndpointBindings(p.Bindings[0], e0)
	_ds.ApplySource(fmt.Sprintf("[Plink][%v]", p.Id))
	ds.Merge(&_ds)
	_ds = CheckEndpointBindings(p.Bindings[1], e1)
	_ds.ApplySource(fmt.Sprintf("[Plink][%v]", p.Id))
	ds.Merge(&_ds)

	return ds
}

func CheckEndpointBindings(bindings string, endpoint addie.Identify) Diagnostics {

	var ds Diagnostics

	bindings = strings.Replace(strings.TrimSuffix(bindings, ","), " ", "", -1)
	bs := strings.Split(bindings, ",")

	switch endpoint.(type) {
	case addie.Sax:
		s := endpoint.(addie.Sax)
		_ds := CheckSaxBindings(bs, s)
		ds.Merge(&_ds)
	}

	return ds
}

func ExtractSensorData(s addie.Sax) (map[string]int, Diagnostics) {

	var ds Diagnostics
	result := make(map[string]int)

	re, _ := regexp.Compile("([a-zA-Z_][a-zA-Z0-9_]*)\\(([0-9]*)\\)")
	ms := re.FindAllStringSubmatch(s.Sense, -1)

	for _, m := range ms {
		i, err := strconv.Atoi(m[2])

		if err != nil {
			ds.Elements = append(ds.Elements,
				Diagnostic{"error",
					fmt.Sprintf("$source "+
						"The sensor rate for sax [%v] binding [%s] = [%s] is not valid, "+
						"it must be an int", s.Id, m[1], m[2])})
		}
		result[m[1]] = i
	}

	return result, ds
}

type Limits struct {
	Static, Dynamic float64
}

func ExtractActuatorData(s addie.Sax) (map[string]Limits, Diagnostics) {

	var ds Diagnostics
	result := make(map[string]Limits)

	re, _ := regexp.Compile(
		"([a-zA-Z_][a-zA-Z0-9_]*)\\(([0-9]+(?:\\.[0-9]+)?),([0-9]+(?:\\.[0-9]+)?)\\)")
	ms := re.FindAllStringSubmatch(strings.Replace(s.Actuate, " ", "", -1), -1)

	for _, m := range ms {
		sval, err := strconv.ParseFloat(m[2], 64)
		if err != nil {
			ds.Elements = append(ds.Elements,
				Diagnostic{"error",
					fmt.Sprintf("$source "+
						"The actuator static limit for sax [%v] binding [%s] = [%s] is not valid, "+
						"it must be an float", s.Id, m[1], m[2])})

		}
		dval, err := strconv.ParseFloat(m[3], 64)
		if err != nil {
			ds.Elements = append(ds.Elements,
				Diagnostic{"error",
					fmt.Sprintf("$source "+
						"The actuator dynamic limit for sax [%v] binding [%s] = [%s] is not valid, "+
						"it must be an float", s.Id, m[1], m[3])})

		}
		result[m[1]] = Limits{sval, dval}
	}

	return result, ds

}

func CheckSaxBindings(bs []string, s addie.Sax) Diagnostics {

	var ds Diagnostics

	sd, _ds := ExtractSensorData(s)
	ds.Merge(&_ds)
	if ds.Fatal() {
		return ds
	}

	ad, _ds := ExtractActuatorData(s)
	ds.Merge(&_ds)
	if ds.Fatal() {
		return ds
	}

	for _, b := range bs {
		_, sok := sd[b]
		_, aok := ad[b]
		if !sok && !aok {
			ds.Elements = append(ds.Elements,
				Diagnostic{"error",
					fmt.Sprintf("$source The binding [%s] does not exist in Sax [%v]",
						b, s.Id)})
		}
	}

	return ds

}
