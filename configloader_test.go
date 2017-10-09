package main

import (
	"fmt"
	"testing"
)

func TestLoadFile(t *testing.T) {
	cfg, err := LoadFile("backupconf.yaml")

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(*cfg)
		// for i, job := range (*cfg).JobConfigs {
		// 	fmt.Printf("%d: %v \n", i, job)
		// }
	}

}
