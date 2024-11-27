package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
)

type ConfigFile struct {
	Servers  []ConfigFileServer `json:"servers"`
	Balancer ConfigFileBalancer `json:"balancer"`
}

type ConfigFileServer struct {
	Port string `json:"port"`
	Name string `json:"name"`
}

type ConfigFileBalancer struct {
	Type    string             `json:"type"`
	Port    string             `json:"port"`
	Name    string             `json:"name"`
	Servers []ConfigFileServer `json:"servers"`
}

//go:embed config.json
var configFile []byte

func main() {
	log.Println("Starting load balancer system...")

	config, err := ParseConfigFile(&configFile)

	if err != nil {
		log.Fatalf("Error parsing config file: %v\n", err)
	}

	servers := make([]Server, len(config.Servers))
	for i, server := range config.Servers {
		servers[i] = Server{
			port: fmt.Sprintf(":%s", server.Port),
			name: server.Name,
		}
	}

	// Technically, this shouldn't be done here, I am doing it here for testing duing development.
	go startServers(servers)

	// TODO:: Add config support for load balancer
	lb := NewLoadBalancer(LoadBalancer{
		name:          config.Balancer.Name,
		port:          fmt.Sprintf(":%s", config.Balancer.Port),
		servers:       &servers,
		balancingAlgo: &RoundRobinBalancer{},
	})

	startLoadBalancer(lb)
}

func ParseConfigFile(file *[]byte) (ConfigFile, error) {
	var config ConfigFile

	err := json.Unmarshal(*file, &config)

	if err != nil {
		return ConfigFile{}, err
	}

	return config, nil
}
