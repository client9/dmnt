package main

import (
	"flag"
	"fmt"
	"github.com/client9/dmnt"
	"log"
	"strings"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		log.Fatalf("Expected at least one arg")
	}
	outflags := make([]string, 0, 2*len(args))
	for _, path := range args {
		outflags = append(outflags, "-v")
		outflags = append(outflags, dmt.ComputeMount(path))
	}
	fmt.Println(strings.Join(outflags, " "))
}
