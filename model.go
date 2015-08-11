/*
This file contains the Cypress data model as perceived by addie
*/
package addie

import (
	"fmt"
)

type Id struct {
	Name   string `json:"name"`
	Sys    string `json:"sys"`
	Design string `json:"design"`
}

type Position struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

type Identify interface {
	Identify() Id
}

//Cyber------------------------------------------------------------------------

type NetHost struct {
	Id
	Interfaces map[string]Interface `json:"interfaces"`
}

func (h *NetHost) Equals(x *NetHost) bool {

	if !(h.Id == x.Id) {
		return false
	}

	for k, v := range h.Interfaces {
		_v, ok := x.Interfaces[k]
		if !ok {
			return false
		}
		if v != _v {
			return false
		}
	}

	return true
}

type Interface struct {
	Name string `json:"name"`
	PacketConductor
}

type Computer struct {
	NetHost
	Position     Position `json:"position"`
	OS           string   `json:"os"`
	Start_script string   `start_script:"start_script"`
}

func (c Computer) Identify() Id { return c.Id }

func (c *Computer) Equals(x *Computer) bool {

	return c.NetHost.Equals(&x.NetHost) &&
		c.Position == x.Position &&
		c.OS == x.OS &&
		c.Start_script == x.Start_script

}

type PacketConductor struct {
	Capacity int `json:"capacity"`
	Latency  int `json:"latency"`
}

type Switch struct {
	NetHost
	PacketConductor
	Position Position `json:"position"`
}

func (s Switch) Identify() Id { return s.Id }

func (s *Switch) Equals(x *Switch) bool {

	return s.NetHost.Equals(&x.NetHost) &&
		s.Position == x.Position &&
		s.PacketConductor == x.PacketConductor

}

type Router struct {
	NetHost
	PacketConductor
	Position Position `json:"position"`
}

func (r Router) Identify() Id { return r.Id }

func (r *Router) Equals(x *Router) bool {
	return r.NetHost.Equals(&x.NetHost) &&
		r.PacketConductor == x.PacketConductor &&
		r.Position == x.Position
}

type NetIfRef struct {
	Id
	IfName string `json:"ifname"`
}

type Link struct {
	Id
	Path []Position `json:"path"`
	PacketConductor
	Endpoints [2]NetIfRef `json:"endpoints"`
}

func (l Link) Identify() Id { return l.Id }

func (l *Link) Equals(x *Link) bool {

	return l.Id == x.Id &&
		l.PacketConductor == x.PacketConductor &&
		l.Endpoints == x.Endpoints

}

//Physical---------------------------------------------------------------------

type Model struct {
	Name      string `json:"name"`
	Equations string `json:"equations"`
}

type Phyo struct {
	Id
	Position Position `json:"position"`
	Model    string   `json:"model"`
	Params   string   `json:"params"`
}

func (m Phyo) Identify() Id { return m.Id }

type Binding [2]string

type Plink struct {
	Id
	Endpoints [2]Id `json:"endpoints"`
	//Bindings  []Binding `json:"bindings"` //derive this after we get the bits
	Bindings [2]string `json:"bindings"`
}

func (p Plink) Identify() Id { return p.Id }

//Cyber-Physical---------------------------------------------------------------

type Target struct {
	Id
	Value string `json:"value"`
}

type Sensor struct {
	Id
	Position Position `json:"position"`
	Target   Target   `json:"target"`
	Rate     uint     `json:"rate"`
}

func (s Sensor) Identify() Id { return s.Id }

type Bound struct{ Min, Max float64 }
type Actuator struct {
	Id
	Position     Position
	Target       Target `json:"target"`
	StaticLimit  Bound  `json:"static_limit"`
	DynamicLimit Bound  `json:"dynamic_limit"`
}

func (a Actuator) Identify() Id { return a.Id }

type Sax struct {
	NetHost
	Position Position `json:"position"`
	Sense    string   `json:"sense"`
	Actuate  string   `json:"actuate"`
}

func (s Sax) Identify() Id { return s.Id }

type Design struct {
	Name     string          `json:"name"`
	Elements map[Id]Identify `json:"elements"`
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
