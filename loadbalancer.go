package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

var HOST = "http://127.0.0.1"

type LoadBalancer struct {
	port          string
	name          string
	servers       *[]Server
	balancingAlgo BalancingAlgorithm
}

type BalancingAlgorithm interface {
	getServer(*[]Server) (string, error)
}

type RoundRobinBalancer struct {
	currentIdx int32
}

// TODO:: implement weighted round robin algo or least connections algo

func NewLoadBalancer(opts LoadBalancer) *LoadBalancer {
	return &LoadBalancer{
		name:          opts.name,
		port:          opts.port,
		servers:       opts.servers,
		balancingAlgo: opts.balancingAlgo,
	}
}

func startLoadBalancer(lb *LoadBalancer) {
	log.Printf("Initializing load balancer %s on port %s\n", lb.name, lb.port)

	if err := lb.Start(); err != nil {
		log.Fatalf("Error starting load balancer: %v\n", err)
	}

}

func (lb *LoadBalancer) balance(w http.ResponseWriter, r *http.Request) {
	serverUrl, err := lb.getServer(lb.balancingAlgo)
	if err != nil {
		log.Printf("Error getting next server: %v\n", err)
		http.Error(w, "No server available", http.StatusServiceUnavailable)
		return
	}

	rp, err := NewProxy(serverUrl)
	if err != nil {
		log.Printf("Error creating proxy: %v\n", err)
		http.Error(w, "Proxy error", http.StatusInternalServerError)
		return
	}

	log.Printf("Forwarding request to server: %s\n", serverUrl)
	w.Header().Set("X-Server-Name", lb.name)
	rp.ServeHTTP(w, r)
}

func (lb *LoadBalancer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		lb.balance(w, r)
	})

	log.Printf("Load balancer %s starting on port %s\n", lb.name, lb.port)
	return http.ListenAndServe(lb.port, mux)
}

func (rb *RoundRobinBalancer) getServer(servers *[]Server) (string, error) {
	current := atomic.AddInt32(&rb.currentIdx, 1)
	idx := int(current) % len(*servers)
	return HOST + (*servers)[idx].port, nil
}

func (lb *LoadBalancer) getServer(algo BalancingAlgorithm) (string, error) {
	return algo.getServer(lb.servers)
}

func NewProxy(target string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	return httputil.NewSingleHostReverseProxy(url), nil
}
