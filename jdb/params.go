package jdb

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/file"
	"github.com/cgalvisleon/elvis/strs"
)

func UpdateParamdb(path, atrib, value string) (et.Item, error) {
	if !file.ExistPath(path) {
		return et.Item{}, errors.New("file not found")
	}

	file, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if err != nil {
		return et.Item{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		vals := strings.Split(line, " ")
		for _, val := range vals {
			val = strings.ReplaceAll(val, "#", "")

			if val == atrib {
				line = atrib + " = " + value
				break
			}
		}
		lines = append(lines, line)
	}

	file.Seek(0, 0)
	file.Truncate(0)

	for _, line := range lines {
		fmt.Fprintln(file, line)
	}

	return et.Item{
		Ok: true,
		Result: et.Json{
			"message": strs.Format(`Update atrib - %s`, atrib),
		},
	}, nil
}
