package crudl

import (
	"os"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"

	gomHTTP "github.com/hauxe/gom/http"
	gomMySQL "github.com/hauxe/gom/mysql"
)

var mu sync.Mutex
var sampleServers []*gomHTTP.Server
var sampleDB *sqlx.DB

func CreateSampleServer(routes ...gomHTTP.ServerRoute) (*gomHTTP.Server, error) {
	mu.Lock()
	defer mu.Unlock()
	host := "localhost"
	port := 1234
	if len(sampleServers) > 0 {
		lastServer := sampleServers[len(sampleServers)-1]
		port = lastServer.Config.Port + 1
	}
	server, err := gomHTTP.CreateServer()
	if err != nil {
		return nil, err
	}
	err = server.Start(server.SetHostPortOption(host, port), server.SetHandlerOption(routes...))
	if err != nil {
		return nil, err
	}
	sampleServers = append(sampleServers, server)
	return server, nil
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
		server.Stop()
	}
	os.Exit(code)
}
