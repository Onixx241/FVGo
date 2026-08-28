package main

import (
	"flag"
	"fmt"
	"log"
	"main/slang"
	"os"
)

func main() {

	fmt.Println("-------------------------------")
	fmt.Println("-----FVGo Frontend for Z3------")

	var filevar string

	flag.StringVar(&filevar, "file", "", "")

	flag.Parse()

	if filevar != "" {

		if slang.CheckIfSlangExists() {

			fmt.Println("Slang exists. Parsing file")

			if slang.GetFileAST(filevar) {

				slang.GetASTJson()
				slang.DecodeJson()
			}

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
