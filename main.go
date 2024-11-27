package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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

func main() {
	log.Println("Starting load balancer system...")

	config, err := ParseConfigFile("config.json")

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
		name:          "Load Balancer 1",
		port:          fmt.Sprintf(":%d", 8080),
		servers:       &servers,
		balancingAlgo: &RoundRobinBalancer{},
	})

	startLoadBalancer(lb)
}

func ParseConfigFile(fileName string) (ConfigFile, error) {
	var config ConfigFile

	file, err := os.ReadFile(fileName)
	if err != nil {
		return ConfigFile{}, err
	}

	err = json.Unmarshal(file, &config)

	if err != nil {
		return ConfigFile{}, err
	}

	return config, nil
}
