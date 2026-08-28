package slang

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

func DecodeJson() bool {

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
		return false
	}

	for _, node := range ast.Design.Members {

		for _, bodyinfo := range node.Body.Members {
			fmt.Println(bodyinfo.Name)
			fmt.Println(bodyinfo.Kind)
			fmt.Println(bodyinfo.Type)

			if bodyinfo.Direction != "" {
				fmt.Println(bodyinfo.Direction)
			}

		}

	}

	return true
}
