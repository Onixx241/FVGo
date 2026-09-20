package main

import (
	"flag"
	"fmt"
	"log"
	"main/bmc"
	"main/slang"
	"main/z3_interface"
	"os"
)

const (
	defaultKBound = 10
)

// move z3 import to Z3Interface.go
func main() {

	PrintBanner()

	var filevar string
	var kBound *int

	debug := true

	if debug {

		filevar = "test.sv"

		debugBound := 10

		kBound = &debugBound

	} else {

		flag.StringVar(&filevar, "file", "", "")

		kBound = flag.Int("depth", defaultKBound, "depth for K unrolling")

		flag.Parse()

	}

	if filevar != "" || debug {

		if slang.CheckIfSlangExists() {

			fmt.Println("Slang exists. Parsing file")

			if slang.GetFileAST(filevar) {

				slang.GetASTJson()

				ast := slang.DecodeJson()

				dict, graph := bmc.AstInstance(ast)

				_ = dict

				z3_interface.StateMachine(*kBound, graph)

			}

		}

		cleanUpArtifacts() //maybe add keep artifacts bool for debugging

	}

}

func cleanUpArtifacts() {
	err := os.Remove("dump")
	if err != nil {
		log.Fatal(err)
	}
}

func PrintBanner() {

	fmt.Println("-------------------------------------------------------------------------")
	fmt.Println("---------------------------FVGo Frontend for Z3--------------------------")
	fmt.Println("-------------------------------------------------------------------------")

}
