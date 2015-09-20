package main

import (
	"fmt"
	"github.com/cycps/xptools/dnsc"
	"os"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: keygen <dns name>\n")
		os.Exit(1)
	}

	key, err := dnsc.Keygen(os.Args[1])
	if err != nil {
		fmt.Fprint(os.Stderr, "error executing dnssec-keygen\n")
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	fmt.Printf("key: %s\n", key)
}
