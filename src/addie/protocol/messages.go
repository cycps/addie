package protocol

import (
	"addie"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type ElementUpdate struct {
	OID     addie.Id
	Type    string
	Element json.RawMessage
}

type ModelUpdate struct {
	OldName string
	Model   addie.Model
}

type Update struct {
	Elements []ElementUpdate
	Models   []ModelUpdate
}

type ElementDelete struct {
	Type    string          `json:"type"`
	Element json.RawMessage `json:"element"`
}

type Delete struct {
	Elements []ElementDelete
}

type NewXP struct {
	Name string
}

type UserDesigns struct {
	Designs []string `json:"designs"`
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
