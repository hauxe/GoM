package trace

import (
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

var mu sync.Mutex
var sampleServers []*httptest.Server

type ServerRoute struct {
	Path    string
	Handler http.HandlerFunc
}

func CreateSampleServer(routes ...ServerRoute) *httptest.Server {
	mux := http.NewServeMux()
	for _, route := range routes {
		if route.Handler != nil {
			mux.HandleFunc(route.Path, route.Handler)
		}
	}
	server := httptest.NewServer(mux)
	mu.Lock()
	defer mu.Unlock()
	sampleServers = append(sampleServers, server)
	return server
}

func TestMain(m *testing.M) {
	code := m.Run()
	// close all sample servers
	for _, server := range sampleServers {
		server.Close()
	}
	os.Exit(code)
}
