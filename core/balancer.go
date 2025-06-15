package core

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"request_balancer/types"
	"strings"
	"time"
)

type BalancerStrategy interface {
	GetNextServer() (*types.Server, error)
}

type LoadBalancer struct {
	strategy BalancerStrategy
}

func NewLoadBalancer(strategy BalancerStrategy) *LoadBalancer {
	return &LoadBalancer{
		strategy: strategy,
	}
}

func (lb *LoadBalancer) Serve(w http.ResponseWriter, r *http.Request) {
	server, err := lb.strategy.GetNextServer()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	targetURL, err := url.Parse(server.URL)
	if err != nil {
		http.Error(w, "invalid backend URL", http.StatusInternalServerError)
		return
	}

	targetPath := strings.TrimRight(targetURL.String(), "/") + r.URL.Path
	if r.URL.RawQuery != "" {
		targetPath += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequest(r.Method, targetPath, r.Body)
	if err != nil {
		http.Error(w, "failed to create request to backend", http.StatusInternalServerError)
		return
	}

	for k, value := range r.Header {
		for _, v := range value {
			req.Header.Add(k, v)
		}
	}

	req.Header.Set("X-forwarded-for", r.RemoteAddr)
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "failed to reach backend server", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	byteResp, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read response body", http.StatusInternalServerError)
		return
	}

	for k, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(k, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	fmt.Fprint(w, string(byteResp))
}
