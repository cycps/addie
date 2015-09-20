package dnsc

import (
	"fmt"
	"io/ioutil"
	"os"
	"text/template"
)

func CopyFile(lfn, templateDir string) {
	fn := templateDir + "/" + lfn
	f, err := ioutil.ReadFile(fn)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"failed to read %s template (from GOPATH)\n", fn)
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	ioutil.WriteFile(lfn, f, 0644)
}

func ApplyTemplate(lfn, ofn, templateDir string, rspec interface{}) {

	f, _ := ioutil.ReadFile(templateDir + "/" + lfn)
	tmpl, _ := template.New(lfn).Parse(string(f))
	of, _ := os.Create(ofn)
	tmpl.Execute(of, rspec)
	of.Close()

}
