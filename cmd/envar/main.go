package main

import (
	"fmt"
	"runtime"
)

func main() {
	if runtime.GOOS == "windows" {
		fmt.Println("This program is not aplicable for windows machine.")
	}
}
