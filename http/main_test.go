package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var sampleServers []*httptest.Server

func CreateSampleServer(routes ...ServerRoute) *httptest.Server {
	mux := http.NewServeMux()
	for _, route := range routes {
		if route.Handler != nil {
			mux.HandleFunc(route.Path, route.Handler)
		}
	}
	server := httptest.NewServer(mux)
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
