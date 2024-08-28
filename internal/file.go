package internal

import (
	"fmt"
	"os"

	"github.com/shirou/gopsutil/v4/disk"
)

func GetAllFile(path string) []string {
	files, err := os.ReadDir(path)
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

func GetRemainingSpace(path string) (uint64, error) {
	us, err := disk.Usage(path)
	if err != nil {
		return 0, err
	}

	return us.Free, nil
}
