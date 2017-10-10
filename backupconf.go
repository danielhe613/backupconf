/*
This package is used to automatically backup network devices' configuration by ssh.
*/
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {

	configFile := flag.String("config.file", "backupconf.yaml", "--config.file=<configuration file name>")

	fmt.Println("To be continued")

	backup, err := LoadFromFile(*configFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	backup.execute()

}
