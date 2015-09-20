package main

import (
	"encoding/json"
	"fmt"
	"github.com/cycps/xptools/dnsc"
	"io/ioutil"
	"os"
)

var templateDir string
var rspec dnsc.ServerSpec

func namedConf() {
	dnsc.CopyFile("named.conf", templateDir)
}

func namedConfLocal() {
	dnsc.ApplyTemplate("named.conf.local", "named.conf.local", templateDir, rspec)
}

func namedConfOptions() {
	dnsc.CopyFile("named.conf.options", templateDir)
}

func keys() {
	dnsc.ApplyTemplate("keys.conf", "keys.conf", templateDir, rspec)
}

func apparmor() {
	dnsc.CopyFile("usr.sbin.named", templateDir)
}

func zoneConf() {
	dnsc.ApplyTemplate(
		"db.__XPNAME__.cypress.net",
		fmt.Sprintf("db.%s.cypress.net", rspec.Xpname),
		templateDir,
		rspec)
}

func setupScript() {
	fn := "setup_dns.sh"
	dnsc.ApplyTemplate(fn, fn, templateDir, rspec)
	os.Chmod(fn, 0755)
}

func main() {

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: dns_config <router spec>\n")
		os.Exit(1)
	}

	fmt.Printf("dns_config v0.1\n\n")

	templateDir = os.Getenv("GOPATH") +
		"/src/github.com/cycps/xptools/dnsc/server_templates"

	fmt.Printf("template dir: %s\n", templateDir)

	rspec_src, _ := ioutil.ReadFile(os.Args[1])
	err := json.Unmarshal(rspec_src, &rspec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", "bad rspec file")
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	namedConf()
	namedConfLocal()
	namedConfOptions()
	keys()
	zoneConf()
	//apparmor()
	setupScript()

}
