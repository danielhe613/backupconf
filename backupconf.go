/*
This package is used to automatically backup network devices' configuration by ssh.
*/
package main

import (
	"fmt"
	"os"
)

func main() {

	fmt.Println("To be continued")

	backup, err := LoadFromFile("backupconf.yaml")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	backup.execute()

}
