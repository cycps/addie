package addie

import (
	"fmt"
)

type Id struct{ Name, Sys, Design string }

type Identify interface {
	Identify() Id
}

//Cyber------------------------------------------------------------------------

type NetHost struct {
	Id
	Interfaces map[string]Interface
}
type Interface struct {
	Name string
	PacketConductor
}

type Computer struct {
	NetHost
	OS           string
	Start_script string
}

func (c Computer) Identify() Id { return c.Id }

type PacketConductor struct {
	Capacity int
	Latency  int
}
type Switch struct {
	Id
	PacketConductor
}
type Router struct {
	NetHost
	PacketConductor
}
type NetIfRef struct {
	Id
	IfName string
}
type Link struct {
	Name string
	PacketConductor
	Endpoints [2]NetIfRef
}

//Physical---------------------------------------------------------------------

type Model struct {
	Id
	Equations []string
}
type VarRef struct {
	Id
	Variable string
}
type Equality struct {
	Id
	lhs, rhs VarRef
}

//Cyber-Physical---------------------------------------------------------------

type Sensor struct {
	Id
	Target VarRef
	Rate   uint
}
type Bound struct{ Min, Max float64 }
type Actuator struct {
	Id
	Target       VarRef
	StaticLimit  Bound
	DynamicLimit Bound
}

/*
type Design struct {
	Name       string
	Computers  map[Id]Computer
	Switches   map[Id]Switch
	Routers    map[Id]Router
	Links      map[Id]Link
	Models     map[Id]Model
	Equalities map[Id]VarRef
	Sensors    map[Id]Sensor
	Actuators  map[Id]Actuator
}
*/

type Design struct {
	Name     string
	Elements map[Id]Identify
}

/*
func (d *Design) String() string {
	s := d.Name + "\n"

	s += "  computers:\n"
	for _, x := range d.Computers {
		s += fmt.Sprint("    ", x, "\n")
	}

	s += "  switches:\n"
	for _, x := range d.Switches {
		s += fmt.Sprint("    ", x, "\n")
	}

	return s
}
*/

func (d *Design) String() string {
	s := d.Name + "\n"

	s += " elements:\n"
	for _, x := range d.Elements {
		s += fmt.Sprint("    ", x, "\n")
	}

	return s
}

/*
func EmptyDesign(name string) Design {
	var m Design
	m.Name = name
	m.Computers = make(map[Id]Computer)
	m.Switches = make(map[Id]Switch)
	m.Routers = make(map[Id]Router)
	m.Links = make(map[Id]Link)
	return m
}
*/

func EmptyDesign(name string) Design {
	var m Design
	m.Name = name
	m.Elements = make(map[Id]Identify)
	return m
}
