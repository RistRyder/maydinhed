package raft

import (
	"log"
	"net"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/ristryder/maydinhed/mapping"
)

type HttpListener[K mapping.ClusteredMarkerKey] struct {
	address     string
	httpHandler *HttpHandler[K]
	listener    net.Listener
}

func NewHttpListener[K mapping.ClusteredMarkerKey](address string, markerCluster mapping.MarkerCluster[K], nodeConfiguration NodeConfiguration, store StoreNode[K]) (*HttpListener[K], error) {
	httpHandler := NewHttpHandler(markerCluster, nodeConfiguration, store)

	return &HttpListener[K]{
		address:     address,
		httpHandler: httpHandler,
	}, nil
}

func (l *HttpListener[K]) Close() error {
	return l.listener.Close()
}

func (l *HttpListener[K]) Start() error {

	httpServer := http.Server{
		Handler: l.httpHandler,
	}

	tcpListener, tcpListenerErr := net.Listen("tcp", l.address)
	if tcpListenerErr != nil {
		return errors.Wrap(tcpListenerErr, "failed to start HTTP server")
	}

	l.listener = tcpListener

	http.Handle("/", l.httpHandler)

	go func() {
		serveErr := httpServer.Serve(l.listener)
		if serveErr != nil {
			log.Fatalf("error with HTTP serve: %s", serveErr)
		}
	}()

	return nil
}
