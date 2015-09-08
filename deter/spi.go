package deter

import (
	"encoding/xml"
	"fmt"
	"github.com/cycps/addie"
	"github.com/deter-project/go-spi/spi"
	"log"
	"reflect"
)

func compComp(c *addie.Computer) spi.Computer {

	var _c spi.Computer
	_c.Name = c.Name
	_c.OSs = []spi.OS{spi.OS{Name: c.OS, Version: ""}}
	_c.Attributes = []spi.TopDLAttribute{
		spi.TopDLAttribute{
			Attribute: "containers:openvz_template",
			Value:     "ubuntu-14.04-x86_64",
		},
		spi.TopDLAttribute{
			Attribute: "osid",
			Value:     "Ubuntu1404-64-STD",
		},
		spi.TopDLAttribute{
			Attribute: "startup",
			Value:     c.Start_script,
		},
	}

	for _, i := range c.Interfaces {
		_c.Interfaces = append(_c.Interfaces,
			spi.Interface{
				Name:      i.Name,
				Substrate: "TODO", //this gets resolved when the links get added
				Capacity:  spi.Capacity{float64(i.Capacity), spi.Kind{"max"}},
				Latency:   spi.Latency{float64(i.Latency), spi.Kind{"max"}},
			},
		)
	}

	return _c

}

var kryCount int

func simComp() spi.Computer {

	var c spi.Computer
	c.Name = fmt.Sprintf("kry%d", kryCount)
	kryCount++
	c.OSs = []spi.OS{spi.OS{Name: "Ubuntu1404-64-STD", Version: ""}}
	c.Attributes = []spi.TopDLAttribute{
		spi.TopDLAttribute{
			Attribute: "containers:openvz_template",
			Value:     "ubuntu-14.04-x86_64",
		},
		spi.TopDLAttribute{
			Attribute: "osid",
			Value:     "Ubuntu1404-64-STD",
		},
		spi.TopDLAttribute{
			Attribute: "startup",
			Value:     "sim_init",
		},
	}

	c.Interfaces = append(c.Interfaces,
		spi.Interface{
			Name:      "eth0",
			Substrate: "krynet",
			Capacity:  spi.Capacity{1000.0, spi.Kind{"max"}},
			Latency:   spi.Latency{0, spi.Kind{"max"}},
		},
	)

	return c

}

func saxComp(s *addie.Sax) spi.Computer {

	var c spi.Computer
	c.Name = s.Name
	c.OSs = []spi.OS{spi.OS{Name: "Ubuntu1404-64-STD", Version: ""}}
	c.Attributes = []spi.TopDLAttribute{
		spi.TopDLAttribute{
			Attribute: "containers:openvz_template",
			Value:     "ubuntu-14.04-x86_64",
		},
		spi.TopDLAttribute{
			Attribute: "osid",
			Value:     "Ubuntu1404-64-STD",
		},
		spi.TopDLAttribute{
			Attribute: "startup",
			Value:     "sax_init",
		},
	}

	c.Interfaces = append(c.Interfaces,
		spi.Interface{
			Name:      "eth0",
			Substrate: "krynet",
			Capacity:  spi.Capacity{1000.0, spi.Kind{"max"}},
			Latency:   spi.Latency{0, spi.Kind{"max"}},
		},
	)

	for _, i := range s.Interfaces {
		c.Interfaces = append(c.Interfaces,
			spi.Interface{
				Name:      i.Name,
				Substrate: "TODO", //this gets resolved when the links get added
				Capacity:  spi.Capacity{float64(i.Capacity), spi.Kind{"max"}},
				Latency:   spi.Latency{float64(i.Latency), spi.Kind{"max"}},
			},
		)
	}

	return c

}

func rtrComp(r *addie.Router) spi.Computer {

	var c spi.Computer
	c.Name = r.Name
	c.OSs = []spi.OS{spi.OS{Name: "Ubuntu Click", Version: "Router"}}
	c.Attributes = []spi.TopDLAttribute{
		spi.TopDLAttribute{
			Attribute: "containers:openvz_template",
			Value:     "ubuntu-14.04-x86_64",
		},
		spi.TopDLAttribute{
			Attribute: "osid",
			Value:     "Ubuntu1404-64-STD",
		},
		spi.TopDLAttribute{
			Attribute: "startup",
			Value:     "router_init",
		},
	}

	for _, i := range r.Interfaces {
		c.Interfaces = append(c.Interfaces,
			spi.Interface{
				Name:      i.Name,
				Substrate: "TODO", //this gets resolved when the links get added
				Capacity:  spi.Capacity{float64(i.Capacity), spi.Kind{"max"}},
				Latency:   spi.Latency{float64(i.Latency), spi.Kind{"max"}},
			},
		)
	}

	return c

}

func swSubstrate(sw *addie.Switch) spi.Substrate {

	var ss spi.Substrate
	ss.Name = sw.Name
	ss.Capacity = spi.Capacity{float64(sw.Capacity), spi.Kind{"max"}}
	ss.Latency = spi.Latency{float64(sw.Latency), spi.Kind{"max"}}

	return ss

}

func krySubstrate() spi.Substrate {

	var ks spi.Substrate
	ks.Name = "krynet"
	ks.Capacity = spi.Capacity{float64(10000.0), spi.Kind{"max"}}
	ks.Latency = spi.Latency{float64(0.0), spi.Kind{"max"}}

	return ks

}

func updateIfxSubstrate(ifxName, ssName string, c *spi.Computer) {

	for i, _ := range c.Interfaces {
		if c.Interfaces[i].Name == ifxName {
			c.Interfaces[i].Substrate = ssName
			break
		}
	}

}

func updateEndpoint(ssName string, ifr addie.NetIfRef, dsg *addie.Design,
	xp *spi.Experiment,
	cMap map[addie.Id]*spi.Computer) {

	e, ok := dsg.Elements[ifr.Id]

	if !ok {
		log.Printf("link '%s' references element '%v' but no such element exists",
			ssName, ifr.Id)
		return
	}

	switch e.(type) {
	case addie.Computer:
		c := e.(addie.Computer)
		updateIfxSubstrate(ifr.IfName, ssName, cMap[c.Id])

	case addie.Router:
		r := e.(addie.Router)
		updateIfxSubstrate(ifr.IfName, ssName, cMap[r.Id])

	case addie.Sax:
		s := e.(addie.Sax)
		updateIfxSubstrate(ifr.IfName, ssName, cMap[s.Id])

	default:
		log.Printf("p2p link '%s' references illegal element '%v'", ssName, ifr.Id)
	}

}

func linkSubstrate(link *addie.Link, dsg *addie.Design,
	xp *spi.Experiment,
	cMap map[addie.Id]*spi.Computer,
	sMap map[addie.Id]*spi.Substrate) *spi.Substrate {

	var ss spi.Substrate
	ss.Name = link.Name
	ss.Capacity = spi.Capacity{float64(link.Capacity), spi.Kind{"max"}}
	ss.Latency = spi.Latency{float64(link.Latency), spi.Kind{"max"}}

	a, ok := dsg.Elements[link.Endpoints[0].Id]
	if !ok {
		log.Printf("link '%s' references element '%v' but no such element exists",
			link.Name, link.Endpoints[0].Id)
		return nil
	}
	b, ok := dsg.Elements[link.Endpoints[1].Id]
	if !ok {
		log.Printf("link '%s' references element '%v' but no such element exists",
			link.Name, link.Endpoints[0].Id)
		return nil
	}

	ta := reflect.TypeOf(a).Name()
	tb := reflect.TypeOf(b).Name()

	taIsHost := ta == "Computer" || ta == "Router" || ta == "Sax"
	tbIsHost := tb == "Computer" || tb == "Router" || tb == "Sax"

	if taIsHost && tbIsHost {
		updateEndpoint(link.Name, link.Endpoints[0], dsg, xp, cMap)
		updateEndpoint(link.Name, link.Endpoints[1], dsg, xp, cMap)
		return &ss
	} else if taIsHost && !tbIsHost {
		sw := b.(addie.Switch)
		updateEndpoint(sw.Name, link.Endpoints[0], dsg, xp, cMap)
		return nil
	} else if !taIsHost && tbIsHost {
		sw := a.(addie.Switch)
		updateEndpoint(sw.Name, link.Endpoints[1], dsg, xp, cMap)
		return nil
	}

	return nil
}

func DesignTopDL(dsg *addie.Design) spi.Experiment {

	kryCount = 0

	var xp spi.Experiment

	cMap := make(map[addie.Id]*spi.Computer)
	sMap := make(map[addie.Id]*spi.Substrate)

	var links []*addie.Link

	for _, e := range dsg.Elements {

		switch e.(type) {
		case addie.Computer:
			c := e.(addie.Computer)
			_c := compComp(&c)
			cMap[c.Id] = &_c
			xp.Elements.Elements = append(xp.Elements.Elements, _c)

		case addie.Sax:
			s := e.(addie.Sax)
			c := saxComp(&s)
			cMap[s.Id] = &c
			xp.Elements.Elements = append(xp.Elements.Elements, c)

		case addie.Switch:
			sw := e.(addie.Switch)
			ss := swSubstrate(&sw)
			sMap[sw.Id] = &ss
			xp.Substrates = append(xp.Substrates, ss)

		case addie.Router:
			rtr := e.(addie.Router)
			c := rtrComp(&rtr)
			cMap[rtr.Id] = &c
			xp.Elements.Elements = append(xp.Elements.Elements, c)

		case addie.Link:
			lnk := e.(addie.Link)
			links = append(links, &lnk)
		}

	}

	xp.Elements.Elements = append(xp.Elements.Elements, simComp())
	xp.Substrates = append(xp.Substrates, krySubstrate())

	for _, l := range links {
		_l := linkSubstrate(l, dsg, &xp, cMap, sMap)
		if _l != nil {
			xp.Substrates = append(xp.Substrates, *_l)
		}
	}

	return xp
}

func CreateDeterXP(user string, dsg *addie.Design) error {

	xp := DesignTopDL(dsg)

	//create a session with the DeterLab SPI
	_, err := spi.Login("deterboss", "muffins")
	if err != nil {
		log.Println(err)
		return fmt.Errorf("unable to login as deterboss")
	}

	topdl, err := xml.MarshalIndent(xp, "  ", "  ")
	if err != nil {
		log.Println(err)
		return fmt.Errorf("failed to serialize topology to topdl xml")
	}

	createResponse, err := spi.CreateExperiment(
		user+":"+dsg.Name, "deterboss", string(topdl))

	if err != nil {
		log.Println(err)
		return fmt.Errorf("failed to create experiment")
	}
	log.Printf("createResponse: %v", createResponse)

	return nil
}
