package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

type Server struct {
	url       url.URL
	port      string
	name      string
	isHealthy bool
}

func startServers(servers []Server) {
	var wg sync.WaitGroup

	for k := range servers {
		wg.Add(1)
		go func(server *Server, k int) {
			defer wg.Done()
			if server.name == "" {
				server.name = "Server " + strconv.Itoa(k+1)
			}
			server.create()
		}(&servers[k], k)
	}

	wg.Wait()
}

func (server *Server) create() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Serving from the server: "+server.name)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if server.isHealthy {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, "healthy")
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			io.WriteString(w, "unhealthy")
		}
	})

	srv := &http.Server{
		Addr:    server.port,
		Handler: mux,
	}

	log.Printf("Server %s listening on %s\n", server.name, server.port)
	err := srv.ListenAndServe()

	if err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}
}
