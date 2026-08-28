package slang

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func CheckIfSlangExists() bool {

	found, err := exec.LookPath("./slang.exe")

	if err != nil {

		fmt.Println("ERROR: Could Not Find Slang.exe")
		log.Fatal(err)
		return false
	}

	if found != "" {
		return true
	}

	return false

}

func GetFileAST(filename string) bool { //rename this later

	cmd, out := exec.Command("./slang.exe", filename, "--ast-json", "dump").CombinedOutput()

	println(string(cmd))

	if out != nil {
		log.Fatal(out)
		return false
	}

	return true

}

func GetASTJson() {

	content, err := os.ReadFile("dump")

	if err != nil {

		log.Fatal(content)
		log.Fatal(err)

	}

	//fmt.Println(string(content))

}
