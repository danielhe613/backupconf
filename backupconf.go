/*
This package is used to automatically backup network devices' configuration by ssh.
*/
package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
)

func init() {

	// Only log the warning severity or above.
	logLevel := flag.String("log.level", "info", "--log.level=<panic|fatal|error|warn|info|debug>")
	switch *logLevel {
	case "panic":
		log.SetLevel(log.PanicLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	}

	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr by default.
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

}

func main() {

	logfileName := flag.String("log.file", "backupconf.log", "--log.file=<log file name>")
	configFile := flag.String("config.file", "backupconf.yaml", "--config.file=<configuration file name>")
	flag.Parse()

	//Log initialization
	file, err := os.OpenFile(*logfileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(file)
	}
	defer file.Close()

	defer log.Info("The Network Device Runtime Configuration Backup is Stopped!")

	log.WithFields(log.Fields{
		"traceID": "backupconf",
		"span":    "main",
	}).Info("The Network Device Running Configuration Backup is Started...")

	//Load configuration to initialize the Backup instance.
	backup, err := LoadFromFile(*configFile)
	if err != nil {
		log.Panic(err)
	}

	//Start executing the backup jobs.
	backup.execute()

}
