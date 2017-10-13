/*
This package is used to automatically backup network devices' configuration by ssh.
*/
package main

import (
	"flag"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

func init() {
	flag.Parse()

	// Only log the warning severity or above.
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
	// log.SetOutput(os.Stdout)

	logFilename := "backupconf_" + dateString + ".log"
	var err error
	logFile, err = os.OpenFile(logFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
	}
}

var (
	essTimestamp = time.Now().Format("2006-01-02-15")
	dateString   = time.Now().Format("2006-01-02")
	configFile   = flag.String("config.file", "backupconf.yaml", "-config.file=<configuration file name>")
	logLevel     = flag.String("log.level", "info", "-log.level=<panic|fatal|error|warn|info|debug>")
	logFile      *os.File
)

func main() {

	//Log initialization
	defer logFile.Close()
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
