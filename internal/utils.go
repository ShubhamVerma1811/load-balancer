package internal

import "encoding/json"

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

func ParseConfigFile(file *[]byte) (ConfigFile, error) {
	var config ConfigFile

	err := json.Unmarshal(*file, &config)

	if err != nil {
		return ConfigFile{}, err
	}

	return config, nil
}
