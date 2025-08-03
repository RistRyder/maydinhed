package raft

import (
	"log"
	"net"
	"net/http"

	"github.com/cockroachdb/errors"
)

type HttpListener struct {
	address  string
	listener net.Listener
	store    Store
}

func NewHttpListener(address string, store Store) *HttpListener {
	return &HttpListener{
		address: address,
		store:   store,
	}
}

func (l *HttpListener) Close() error {
	return l.listener.Close()
}

func (l *HttpListener) Start() error {
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
