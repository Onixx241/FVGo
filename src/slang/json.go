package slang

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

func DecodeJson() SlangAST {

	astJson, err := os.Open("dump")

	if err != nil {
		log.Fatal(err)
	}

	defer astJson.Close()

	fileBytes, err := io.ReadAll(astJson)

	if err != nil {
		log.Fatal(err)
	}

	var ast SlangAST

	err = json.Unmarshal(fileBytes, &ast)
	if err != nil {
		log.Fatal(err)
		return SlangAST{}
	}

	return ast
}
