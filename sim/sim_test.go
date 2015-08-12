package sim

import (
	"testing"
)

var expected_src = `
Object Rotor
  w' = tau - H*w^2
	theta' = w

Simulation RotorSim
  Rotor0 rotor0(H:1.5)
  Actuator sax0_A_tau(Min:-10, Max:10, DMin:-0.4, DMax:0.4)
  Sensor sax0_S_w(Rate:30, Destination:localhost)

	sax0_A_tau.u ~ rotor0.tau
	rotor0.w ~ sax0_S_w.y
`

func TestGenerateSourceFromDB(t *testing.T) {

	src, err := GenerateSourceFromDB("chinook", "murphy")
	if err != nil {
		t.Log(err)
		t.Fatal("db failure")
	}

	if src != expected_src {
		t.Log("src:")
		t.Log(src)
		t.Fatal("the generated source is not correct")
	}

}
