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
	Elements []addie.Id
}
