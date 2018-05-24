package crudl

import (
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"

	gomHTTP "github.com/hauxe/GoM/http"
	gomMySQL "github.com/hauxe/GoM/mysql"
)

var mu sync.Mutex
var sampleServers []*httptest.Server
var sampleDB *sqlx.DB

func CreateSampleServer(routes ...gomHTTP.ServerRoute) *httptest.Server {
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
	client, err := gomMySQL.CreateClient()
	if err != nil {
		panic(err)
	}
	err = client.Connect()
	if err != nil {
		panic(err)
	}
	sampleDB = client.C
	defer client.Disconnect()
	code := m.Run()
	// close all sample servers
	for _, server := range sampleServers {
		server.Close()
	}
	os.Exit(code)
}
