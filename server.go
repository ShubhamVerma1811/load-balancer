package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

// TODO:: health checks
type Server struct {
	url  url.URL
	port string
	name string
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

	// TODO:: to figure out how to know active connections for the weighting algo
	srv := &http.Server{
		Addr:    server.port,
		Handler: mux,
		ConnState: func(c net.Conn, cs http.ConnState) {
			//	log.Printf("Connection state: %s\n", cs)
		},
	}

	log.Printf("Server %s listening on %s\n", server.name, server.port)
	err := srv.ListenAndServe()

	if err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}
}

// func logRequestDetails(r *http.Request) {
// 	fmt.Printf("Received request from %s\n", r.RemoteAddr)
// 	fmt.Printf("%s %s %s\n", r.Method, r.URL, r.Proto)
// 	for name, values := range r.Header {
// 		for _, value := range values {
// 			fmt.Printf("%s: %s\n", name, value)
// 		}
// 	}
// }
