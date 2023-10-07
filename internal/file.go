package internal

import (
	"fmt"
	"io/ioutil"
)

func GetAllFile(path string) []string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var filenames []string
	for _, f := range files {
		filenames = append(filenames, f.Name())
	}
	return filenames
}
