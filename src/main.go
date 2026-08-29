package main

import (
	"flag"
	"fmt"
	"log"
	"main/bmc"
	"main/slang"
	"os"
	//"github.com/aclements/go-z3/z3"
)

func main() {

	fmt.Println("-------------------------------------------------------------------------")
	fmt.Println("---------------------------FVGo Frontend for Z3--------------------------")
	fmt.Println("-------------------------------------------------------------------------")

	var filevar string

	debug := false

	if debug {
		filevar = "test.sv"
	} else {
		flag.StringVar(&filevar, "file", "", "")

		flag.Parse()
	}

	if filevar != "" || debug {

		if slang.CheckIfSlangExists() {

			fmt.Println("Slang exists. Parsing file")

			if slang.GetFileAST(filevar) {

				slang.GetASTJson()

				ast := slang.DecodeJson()

				bmc.AstInstance(ast)

			}

		}

		cleanUpArtifacts()

	}

}

func cleanUpArtifacts() {
	err := os.Remove("dump")
	if err != nil {
		log.Fatal(err)
	}
}
