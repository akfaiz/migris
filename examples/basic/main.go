package main

import (
	"log"
	"os"

	"github.com/akfaiz/migris/examples/basic/cmd"
)

func main() {
	if err := cmd.Execute(os.Args); err != nil {
		log.Printf("Error running app: %v\n", err)
		os.Exit(1)
	}
}
