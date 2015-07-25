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

func (s Switch) Identify() Id { return s.Id }

type Router struct {
	NetHost
	PacketConductor
}

func (r Router) Identify() Id { return r.Id }

type NetIfRef struct {
	Id
	IfName string
}

type Link struct {
	Id
	PacketConductor
	Endpoints [2]NetIfRef
}

//Physical---------------------------------------------------------------------

type Model struct {
	Id
	Equations []string
}

func (m Model) Identify() Id { return m.Id }

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

func (s Sensor) Identify() Id { return s.Id }

type Bound struct{ Min, Max float64 }
type Actuator struct {
	Id
	Target       VarRef
	StaticLimit  Bound
	DynamicLimit Bound
}

func (a Actuator) Identify() Id { return a.Id }

type Design struct {
	Name     string
	Elements map[Id]Identify
}

func (d *Design) String() string {
	s := d.Name + "\n"

	s += " elements:\n"
	for _, x := range d.Elements {
		s += fmt.Sprintf("    [%T]: %v \n", x, x)
	}

	return s
}

func EmptyDesign(name string) Design {
	var m Design
	m.Name = name
	m.Elements = make(map[Id]Identify)
	return m
}
