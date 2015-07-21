package addie

type Id struct{ Name, Sys, Design string }

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

type Design struct {
	Name      string
	Computers map[Id]Computer
	Switches  map[Id]Switch
	Routers   map[Id]Router
	Links     map[Id]Link
}

func EmptyDesign(name string) Design {
	var m Design
	m.Name = name
	m.Computers = make(map[Id]Computer)
	m.Switches = make(map[Id]Switch)
	m.Routers = make(map[Id]Router)
	m.Links = make(map[Id]Link)
	return m
}
