package protocol

import (
	"github.com/cycps/addie"
)

type Id struct{ Name, Sys, Design *string }

func (x *Id) Resolve() *addie.Id {
	return nil
}

//Cyber------------------------------------------------------------------------

type NetHost struct {
	Id
	Interfaces *map[string]Interface
}
type Interface struct {
	Name *string
	PacketConductor
}
type Computer struct {
	NetHost
	OS           *string
	Start_script *string
}

type PacketConductor struct {
	Capacity *int
	Latency  *int
}

type UpdateMsg struct {
	Computers *[]Computer
}

type Update struct {
	Computers []addie.Computer
}

func (x *UpdateMsg) Resolve() *UpdateMsg {
	//TODO you are here
	return nil
}

type Delete struct {
	Computers *[]Id
}
