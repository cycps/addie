package addie

import (
	"errors"
	"log"
	"strings"
)

type Element struct {
	Name string
	Sys  string
}

type NetworkHost struct {
	Element
	Interfaces []Interface
}

type Interface struct {
	Name string
}

type Computer struct {
	NetworkHost
	OS           string
	Start_script string
}

type Switch struct {
	Element
	capacity int
	latency  int
}

type Router struct {
	NetworkHost
	capacity int
	latency  int
}

type Link struct {
	Name      string
	capacity  int
	endpoints [2]NetworkHost
}

type WanLink struct {
	Link
	latency int
}

type System struct {
	Element
	Systems   []System
	Computers []Computer
	Switches  []Switch
	Routers   []Router
}

func (s *System) DeleteElement(e Element) error {
	sys := s.FindSubSystem(e.Sys)
	if sys == nil {
		return errors.New("unkown system " + e.Sys)
	}

	found := false
	var idx int
	for i, c := range sys.Computers {
		if c.Name == e.Name {
			found = true
			idx = i
			break
		}
	}
	if found {
		sys.Computers = append(sys.Computers[:idx], sys.Computers[1+idx:]...)
		return nil
	}
	for i, s := range sys.Switches {
		if s.Name == e.Name {
			found = true
			idx = i
			break
		}
	}
	if found {
		sys.Switches = append(sys.Switches[:idx], sys.Switches[1+idx:]...)
		return nil
	}
	for i, r := range sys.Routers {
		if r.Name == e.Name {
			found = true
			idx = i
			break
		}
	}
	if found {
		sys.Routers = append(sys.Routers[:idx], sys.Routers[1+idx:]...)
		return nil
	}
	return nil
}

func (s *System) FindSubSystem(name string) *System {
	log.Printf("[FindSubSystem] `%s` -> `%s`", s.Name, name)
	ss := strings.SplitAfterN(name, ".", 3)
	if ss[0] != s.Name {
		return nil
	}
	switch len(ss) {
	case 1:
		return s
	case 2:
		for i := range s.Systems {
			if s.Systems[i].Name == ss[1] {
				return &s.Systems[i]
			}
		}
	case 3:
		for i := range s.Systems {
			if s.Systems[i].Name == ss[1] {
				return s.FindSubSystem(ss[1] + ss[2])
			}
		}
	}
	return nil
}

func (s *System) FindComputer(e Element) *Computer {
	sys := s.FindSubSystem(e.Sys)
	if sys == nil {
		return nil
	}
	for i := range sys.Computers {
		if sys.Computers[i].Name == e.Name {
			return &sys.Computers[i]
		}
	}
	return nil
}

func (s *System) AddComputer(c Computer) *Computer {
	sys := s.FindSubSystem(c.Sys)
	if sys == nil {
		return nil
	}
	sys.Computers = append(sys.Computers, c)
	_s := &sys.Computers[len(sys.Computers)-1]
	return _s
}
