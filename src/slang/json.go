package slang

import (
	"encoding/json"
	"io"
	"log"
	"main/bmc"
	"os"
)

func DecodeJson() bmc.SlangAST {

	astJson, err := os.Open("dump")

	if err != nil {
		log.Fatal(err)
	}

	defer astJson.Close()

	fileBytes, err := io.ReadAll(astJson)

	if err != nil {
		log.Fatal(err)
	}

	var ast bmc.SlangAST

	err = json.Unmarshal(fileBytes, &ast)
	if err != nil {
		log.Fatal(err)
		return bmc.SlangAST{}
	}

	return ast
}
