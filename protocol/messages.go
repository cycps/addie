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

type Update struct {
	Elements []ElementUpdate
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
