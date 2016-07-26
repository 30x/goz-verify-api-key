package main

import (
	"github.com/30x/gozerian/go_gateway"
	"github.com/30x/gozerian/pipeline"
	"os"
	"fmt"
	"github.com/30x/gozerian/test_util"
)

// This is just an example using go_gateway. Config via main.yaml.

func main() {
	pipeline.RegisterDie("dump", test_util.CreateDumpFitting)

	yamlReader, err := os.Open("main.yaml")
	if err != nil {
		fmt.Print(err)
	}

	err = go_gateway.ListenAndServe(yamlReader)
	if err != nil {
		fmt.Print(err)
	}
}
