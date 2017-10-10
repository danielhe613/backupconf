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
		for _, action := range cfg.Jobs[0].Actions {
			if action.Expect != "" {
				fmt.Println(action.Expect)
			} else if action.Send != "" {
				fmt.Println(action.Send)
			}
		}
	}

}
