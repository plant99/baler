package baler

import (
	"fmt"
	"os"
)

func Convert(path string, limit uint32) error {
	pathFixed := "/Users/shiv/code/experimental/heli-exp"
	entries, err := os.ReadDir(pathFixed)
	if err != nil {
		return err
	}
	// sort entries alphabetically
	fmt.Println(entries)
	return nil
}
