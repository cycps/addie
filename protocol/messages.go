package protocol

import (
	"encoding/json"
	"github.com/cycps/addie"
)

type ElementUpdate struct {
	OID     addie.Id
	Type    string
	Element json.RawMessage
}

/*
type ComputerUpdate struct {
	OID  addie.Id
	Data addie.Computer
}
*/

//func (c ComputerUpdate) Id() addie.Id            { return c.OID }
//func (c ComputerUpdate) Element() addie.Identify { return c.Data }

type Update struct {
	Elements []ElementUpdate
}

/*
type Update struct {
	Computers  []ComputerUpdate
	Switches   []addie.Switch
	Routers    []addie.Router
	Links      []addie.Link
	Models     []addie.Model
	Equalities []addie.VarRef
	Sensors    []addie.Sensor
	Actuators  []addie.Actuator
}
*/

type Delete struct {
	Computers  []addie.Id
	Switches   []addie.Id
	Routers    []addie.Id
	Links      []addie.Id
	Models     []addie.Id
	Equalities []addie.Id
	Sensors    []addie.Id
	Actuators  []addie.Id
}
