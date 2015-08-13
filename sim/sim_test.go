package sim

import (
	"io/ioutil"
	"os"
	"testing"
)

var expected_src = `Object Rotor(H)
  w' = tau - H*w^2
  theta' = w

Simulation chinook
  Rotor rtr(H:2.5)
  Sensor sax0_S_w(Rate:30, Destination:localhost)
  Actuator sax0_A_tau(Min:-10, Max:10, DMin:-0.4, DMax:0.4)

  rtr.w ~ sax0_S_w.y
  rtr.tau ~ sax0_A_tau.u
`

func TestGenerateSourceFromDB(t *testing.T) {

	src, err := GenerateSourceFromDB("chinook", "murphy")
	if err != nil {
		t.Log(err)
		t.Fatal("db failure")
	}

	if src != expected_src {
		t.Log("src:")
		t.Log("\n`" + src + "\n`")
		t.Log("\n`" + expected_src + "\n`")
		t.Fatal("the generated source is not correct")
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Log(err)
		t.Fatal("cwd anger")
	}
	out := wd + "/" + "chinook.cyp"
	t.Log(out)
	err = ioutil.WriteFile(out, []byte(src), 0664)
	if err != nil {
		t.Log(err)
		t.Fatal("could not write sim source")
	}

}
