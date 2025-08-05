package raft

import (
	"log"
	"net"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/ristryder/maydinhed/stores"
)

type HttpListener[K stores.StoreKey] struct {
	address  string
	listener net.Listener
	store    StoreNode[K]
}

func NewHttpListener[K stores.StoreKey](address string, store StoreNode[K]) *HttpListener[K] {
	return &HttpListener[K]{
		address: address,
		store:   store,
	}
}

func (l *HttpListener[K]) Close() error {
	return l.listener.Close()
}

func (l *HttpListener[K]) Start() error {
	httpHandler := NewHttpHandler(l.store)

	httpServer := http.Server{
		Handler: httpHandler,
	}

	tcpListener, tcpListenerErr := net.Listen("tcp", l.address)
	if tcpListenerErr != nil {
		return errors.Wrap(tcpListenerErr, "failed to start HTTP server")
	}

	l.listener = tcpListener

	http.Handle("/", httpHandler)

	go func() {
		serveErr := httpServer.Serve(l.listener)
		if serveErr != nil {
			log.Fatalf("error with HTTP serve: %s", serveErr)
		}
	}()

	return nil
}
