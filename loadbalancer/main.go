package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync/atomic"

	"github.com/ShubhamVerma1811/load-balancer/internal"
)

var HOST = "http://127.0.0.1"

type Server struct {
	Port      string
	Name      string
	IsHealthy bool
}

type LoadBalancer struct {
	Port          string
	Name          string
	Servers       *[]Server
	BalancingAlgo BalancingAlgorithm
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
		Name:          opts.Name,
		Port:          opts.Port,
		Servers:       opts.Servers,
		BalancingAlgo: opts.BalancingAlgo,
	}
}

func StartLoadBalancer(lb *LoadBalancer) {
	log.Printf("Initializing load balancer %s on port %s\n", lb.Name, lb.Port)

	if err := lb.Start(); err != nil {
		log.Fatalf("Error starting load balancer: %v\n", err)
	}

}

func (lb *LoadBalancer) balance(w http.ResponseWriter, r *http.Request) {
	serverUrl, err := lb.getServer(lb.BalancingAlgo)
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
	w.Header().Set("X-Server-Name", lb.Name)
	rp.ServeHTTP(w, r)
}

func (lb *LoadBalancer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		lb.balance(w, r)
	})

	log.Printf("Load balancer %s starting on port %s\n", lb.Name, lb.Port)
	return http.ListenAndServe(lb.Port, mux)
}

func (rb *RoundRobinBalancer) getServer(servers *[]Server) (string, error) {
	current := atomic.AddInt32(&rb.currentIdx, 1)
	idx := int(current) % len(*servers)

	for {
		if (*servers)[idx].IsHealthy {
			break
		}
		idx = (idx + 1) % len(*servers)
		log.Printf("Server %s is not healthy, trying next one\n", (*servers)[idx].Name)
	}

	return HOST + (*servers)[idx].Port, nil
}

func (lb *LoadBalancer) getServer(algo BalancingAlgorithm) (string, error) {
	return algo.getServer(lb.Servers)
}

func NewProxy(target string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	return httputil.NewSingleHostReverseProxy(url), nil
}

func main() {

	configFile, err := os.ReadFile("../config.json")

	log.Println("Starting load balancer system...")

	config, err := internal.ParseConfigFile(&configFile)

	if err != nil {
		log.Fatalf("Error parsing config file: %v\n", err)
	}

	log.Println(config.Balancer.Name, config.Servers[0].Name)

	servers := make([]Server, len(config.Servers))
	for i, s := range config.Servers {
		servers[i] = Server{
			Port:      fmt.Sprintf(":%s", s.Port),
			Name:      s.Name,
			IsHealthy: true,
		}
	}

	lb := NewLoadBalancer(LoadBalancer{
		Name:          "Load Balancer",
		Port:          fmt.Sprintf(":%d", 8080),
		Servers:       &servers,
		BalancingAlgo: &RoundRobinBalancer{},
	})

	StartLoadBalancer(lb)
}
