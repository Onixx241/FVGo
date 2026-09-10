package main

import (
	"flag"
	"fmt"
	"log"
	"main/bmc"
	"main/slang"
	"os"

	z3 "github.com/Z3Prover/z3/src/api/go"
)

func main() {

	PrintBanner()

	var filevar string

	debug := true

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

				dict, graph := bmc.AstInstance(ast)

				_ = dict

				bmc.DumpFlattened(graph)

				bmc.DumpAssertions(graph)

				ctx := z3.NewContext()
				solver := ctx.NewSolver()

				solver.Assert(&z3.Expr{})

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
