package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/ShubhamVerma1811/load-balancer/internal"
)

type Server struct {
	Port      string
	Name      string
	IsHealthy bool
}

func StartServers(servers []Server) {
	var wg sync.WaitGroup

	for k := range servers {
		wg.Add(1)
		go func(server *Server, k int) {
			defer wg.Done()
			if server.Name == "" {
				server.Name = "Server " + strconv.Itoa(k+1)
			}
			server.create()
		}(&servers[k], k)
	}

	wg.Wait()
}

func (server *Server) create() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Serving from the server: "+server.Name)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if server.IsHealthy {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, "healthy")
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			io.WriteString(w, "unhealthy")
		}
	})

	srv := &http.Server{
		Addr:    server.Port,
		Handler: mux,
	}

	log.Printf("Server %s listening on %s\n", server.Name, server.Port)
	err := srv.ListenAndServe()

	if err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}
}

func main() {
	configFile, err := os.ReadFile("../config.json")

	if err != nil {
		log.Fatalf("Error reading config file: %v\n", err)
	}

	config, err := internal.ParseConfigFile(&configFile)

	servers := make([]Server, len(config.Servers))
	for i, s := range config.Servers {
		servers[i] = Server{
			Port:      fmt.Sprintf(":%s", s.Port),
			Name:      s.Name,
			IsHealthy: true,
		}
	}

	StartServers(servers)

}
