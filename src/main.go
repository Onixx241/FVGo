package main

import (
	"fmt"
	"log"
	"main/slang"
	"os"
)

func main() {

	fmt.Println("-------------------------------")
	fmt.Println("-----FVGo Frontend for Z3------")

	if slang.CheckIfSlangExists() {

		fmt.Println("Slang exists. Parsing file")

		if slang.GetFileAST("Splitter.sv") { //replace this with command line arg

			slang.GetASTJson()
			slang.DecodeJson()
		}

	}

	cleanUpArtifact()

}

func cleanUpArtifact() {
	err := os.Remove("dump")
	if err != nil {
		log.Fatal(err)
	}
}
