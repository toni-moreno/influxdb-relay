package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/toni-moreno/influxdb-srelay/backend"
	"github.com/toni-moreno/influxdb-srelay/cluster"
	"github.com/toni-moreno/influxdb-srelay/config"
	"github.com/toni-moreno/influxdb-srelay/relayservice"
	"github.com/toni-moreno/influxdb-srelay/utils"
)

const (
	relayVersion = "0.2.0"
)

var (
	usage = func() {
		fmt.Println("Please, see README for more information about InfluxDB Relay...")
		flag.PrintDefaults()
	}

	configFile  = flag.String("config", "", "Configuration file to use")
	logDir      = flag.String("logdir", "", "Default log Directory")
	verbose     = flag.Bool("v", false, "If set, InfluxDB Relay will log HTTP requests")
	versionFlag = flag.Bool("version", false, "Print current InfluxDB Relay version")
)

func runRelay(cfg config.Config, logdir string) {
	relay, err := relayservice.New(cfg, logdir)
	if err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		<-sigChan
		relay.Stop()
	}()

	log.Println("starting relays...")
	relay.Run()
}

func main() {

	var err error

	flag.Usage = usage
	flag.Parse()

	if *versionFlag {
		fmt.Println("influxdb-srelay version " + relayVersion)
		return
	}

	// Configuration file is mandatory
	if *configFile == "" {
		fmt.Fprintln(os.Stderr, "Missing configuration file")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if len(*logDir) == 0 {
		*logDir, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		//check if exist and
		if _, err := os.Stat(*logDir); os.IsNotExist(err) {
			os.Mkdir(*logDir, 0755)
		}
	}

	// And it has to be loaded in order to continue
	cfg, err := config.LoadConfigFile(*configFile)
	if err != nil {
		log.Println("Version: " + relayVersion)
		log.Fatal(err.Error())
	}
	utils.SetLogdir(*logDir)
	utils.SetVersion(relayVersion)
	backend.SetLogdir(*logDir)
	backend.SetConfig(cfg)
	cluster.SetLogdir(*logDir)
	cluster.SetConfig(cfg)

	runRelay(cfg, *logDir)
}
