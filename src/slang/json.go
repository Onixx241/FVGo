package slang

import (
	"encoding/json"
	"io"
	"log"
	"main/bmc/nodes"
	"os"
)

func DecodeJson() nodes.SlangAST {

	astJson, err := os.Open("dump")

	if err != nil {
		log.Fatal(err)
	}

	defer astJson.Close()

	fileBytes, err := io.ReadAll(astJson)

	if err != nil {
		log.Fatal(err)
	}

	var ast nodes.SlangAST

	err = json.Unmarshal(fileBytes, &ast)
	if err != nil {
		log.Fatal(err)
		return nodes.SlangAST{}
	}

	return ast
}
