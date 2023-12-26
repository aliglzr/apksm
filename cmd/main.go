package main

import (
	"apksm"
	"apksm/logger"
	"flag"
	"log"
	"os"
	"os/user"
	"time"
)

var configDefaultPath = "configs/default.json"
var logTimeFormat = "2006-01-02"
var configPath = flag.String("config", configDefaultPath, "Configuration file path")
var logPath = flag.String("log", "logs/apksm-"+time.Now().Format(logTimeFormat)+".log", "Log file store path")
var address = flag.String("host", "127.0.0.1", "Port for http server")
var port = flag.String("port", "8080", "Address for http server")
var logging = flag.Bool("logging", true, "Enabling or disabling logging to file")
var logfilter = flag.String("logfilter", "", "The text to filter log by")

func isRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("[isRoot] Unable to get current user: %s", err)
	}
	return currentUser.Username == "root"
}

func main() {
	flag.Parse()

	jsonData, err := os.ReadFile(*configPath)
	if err != nil {
		if *configPath == configDefaultPath {
			// If user doesn't pass config path
			panic("Default configuration file does not exist at " + configDefaultPath)
		} else {
			// If user pass config path to program
			panic("Error while reading configuration file from " + *configPath)
		}
	}

	if *logging == false {
		logger.Logln("File logging disabled!")
		logger.Disable()
	}

	if !isRoot() {
		logger.Logln("You must run this program with sudo permission")
		return
	} else {
		logger.Logln("Sudo permission granted!")
	}

	if *logfilter != "" {
		logger.Filter(*logfilter)
	}

	logger.SetFilename(*logPath)

	config := apksm.NewConfig(jsonData)
	monitor := apksm.NewMonitor(config)
	httpAdd := *address + ":" + *port
	logger.Logln("Http server listening on", httpAdd)
	go apksm.RunHttp(httpAdd, monitor)
	monitor.Run()
}
