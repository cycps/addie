package protocol

import (
	"github.com/cycps/addie"
)

type Update struct {
	Computers  []addie.Computer
	Switches   []addie.Switch
	Routers    []addie.Router
	Links      []addie.Link
	Models     []addie.Model
	Equalities []addie.VarRef
	Sensors    []addie.Sensor
	Actuators  []addie.Actuator
}

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
