/*
This package is used to automatically backup network devices' configuration by ssh.
*/
package main

import (
	"fmt"
	"os"
	"time"
)

func main() {

	fmt.Println("To be continued")

	cfg := loadConfig()

	for _, job := range cfg.Jobs {
		executeJob(job, cfg)
	}

}

func loadConfig() *Backup {
	cfg, err := LoadFile("backupconf.yaml")
	if err != nil {
		fmt.Println("Failed to load backupconf.yaml.")
		os.Exit(1)
	}

	timeout, err := time.ParseDuration(cfg.TimeoutStr)
	if err != nil {
		fmt.Println("The timeout defined in global config is invalid!")
		os.Exit(1)
	}
	cfg.timeout = timeout

	return cfg
}

func executeJob(job Job, cfg *Backup) {

	//Clear the old backup configuration first.

	//Controls the device to backup running configuration automatically.

	//Upload the backup configuration to ESS.

}

func executeJobOnTarget(target string, job Job) {

	sshClient, err := NewSSHClient(target, job.Username, "passwordToModify")
	if err != nil {
		fmt.Println("Failed to create SSH Client to target " + target)
		return
	}

	for _, action := range job.Actions {
		if action.Send != "" {
			fmt.Println(action.Send)
			sshClient.Send(action.Send)
		} else if action.Expect != "" {
			fmt.Println(action.Expect)
			sshClient.Expect(action.Expect, time.Duration(5))
		}
	}
}
