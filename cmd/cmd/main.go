package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	file, err := os.ReadFile("./commands.json")
	if err != nil {
		fmt.Println("Error al leer el archivo JSON:", err)
		return
	}

	var commands []string
	err = json.Unmarshal(file, &commands)
	if err != nil {
		fmt.Println("Error al leer el archivo JSON:", err)
		return
	}

	for _, command := range commands {
		exit := ExecCommand(command)
		fmt.Println(string(exit))
	}
}

func ExecCommand(command string) []byte {
	args := strings.Split(command, " ")
	if len(args) == 0 {
		return []byte("Not command")
	}

	cmd := exec.Command(args[0], args[1:]...)
	execute, err := cmd.CombinedOutput()
	if err != nil {
		return []byte(err.Error())
	}

	return execute
}
