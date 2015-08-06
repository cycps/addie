package protocol

import (
	"bytes"
	"encoding/json"
	"github.com/cycps/addie"
	"log"
	"net/http"
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

type NewXP struct {
	Name string
}

func Unpack(r *http.Request, x interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	err := json.Unmarshal(buf.Bytes(), &x)
	if err != nil {
		log.Println("[unpack] bad message")
		log.Println(err)
		log.Println(buf.String())
		return nil
	}
	return nil
}
