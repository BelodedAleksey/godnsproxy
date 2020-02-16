package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
)

var (
	config           *Config
	blackList        *List
	localAliasesList *List
	cache            Cache

	configFile = flag.String("c", "etc/config.yaml", "configuration file")
)

func listenInterrupt() {
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	for {
		select {
		case <-sig:
			log.Println("Terminating...")
			return
		}
	}
}

func changeDNSSettings(config *Config) {
	system := runtime.GOOS
	if system == "windows" {
		if config.Nameservers == nil {
			//Get dns ip
			dnsIP, err := GetDNSIP(config.NetInterfaceName)
			if err != nil {
				fmt.Println("Error getDNSIP: ", err.Error())
			}
			fmt.Println("DNSIP: ", dnsIP)
			config.Nameservers = []string{dnsIP}
		}

		//Set dns = nameServer
		err := SetDNSIP(config.NetInterfaceName, config.LocalDNSProxy)
		if err != nil {
			fmt.Println("Error setDNSIP: ", err.Error())
		}

	} else {
		log.Fatalf("%s is not supported\n", system)
	}
}

func main() {

	flag.Parse()

	var err error
	config, err = loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		//Set dns to automatically by dhcp
		err := SetDNSIP(config.NetInterfaceName, "dhcp")
		if err != nil {
			fmt.Println("Error setDNSIP: ", err.Error())
		}
	}()

	//Change network adapter settings
	changeDNSSettings(config)

	cache = NewMemoryCache()

	blackList = UpdateList(config.Blocklist)
	localAliasesList = UpdateList(config.LocalAliases)

	startServer(config.LocalDNSProxy)
	listenInterrupt()
}
