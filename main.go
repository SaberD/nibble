package main

import (
	"flag"
	"fmt"
)

var version = "dev"

func main() {
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.Parse()

	if showVersion {
		fmt.Println(version)
		return
	}

	fmt.Println("Hello from YOUR_PROJECT!")
}
