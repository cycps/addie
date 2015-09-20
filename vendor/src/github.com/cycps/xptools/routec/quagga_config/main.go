package main

import (
	"fmt"
	"github.com/cycps/xptools/routec"
	"io/ioutil"
	"os"
)

func main() {

	fmt.Printf("quagga_config v0.1\n\n")
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: quagga_config <neighbors json>\n")
		os.Exit(1)
	}

	chart := routec.ReadRouterChart(os.Args[1])
	cfg := routec.InitRouterConfig(&chart)

	routec.ResolveInterfaceInfo(&cfg, &chart)

	cfg.GenOspf6Conf()
	ioutil.WriteFile("ospf6d.conf", []byte(cfg.Ospf6Conf), 0644)

	cfg.GenZebraConf()
	ioutil.WriteFile("zebra.conf", []byte(cfg.ZebraConf), 0644)

}
