package main

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-yaml/yaml"
)

type Config struct {
	NetInterfaceName string        `yaml:"netInterfaceName"`
	LocalDNSProxy    string        `yaml:"localDNSProxy"`
	Nameservers      []string      `yaml:"nameservers"`
	Blocklist        []string      `yaml:"blocklist"`
	LocalAliases     []string      `yaml:"localaliases"`
	BlockAddress4    string        `yaml:"blockAddress4"`
	BlockAddress6    string        `yaml:"blockAddress6"`
	ConfigUpdate     bool          `yaml:"configUpdate"`
	UpdateInterval   time.Duration `yaml:"updateInterval"`
}

func loadConfig() (*Config, error) {
	config := &Config{}

	if _, err := os.Stat(*configFile); err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	if config.ConfigUpdate {
		go configWatcher()
	}
	return config, nil
}

func configWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Add(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("Config file updated, reload config")
				c, err := loadConfig()
				if err != nil {
					log.Println("Bad config: ", err)
				} else {
					log.Println("Config successfuly updated")
					config = c
					if !c.ConfigUpdate {
						return
					}
				}
			}
		case err := <-watcher.Errors:
			log.Println("error:", err)
		}
	}
}
