package raft

import (
	"encoding/json"
	"net/http"

	"github.com/ristryder/maydinhed/stores"
)

type HttpHandler[K stores.StoreKey] struct {
	EnableEgress  bool
	EnableIngress bool
	store         StoreNode[K]
}

func (h *HttpHandler[K]) handleGetLocationRequest(responseWriter http.ResponseWriter, request *http.Request) {
	if !h.EnableEgress {
		responseWriter.WriteHeader(http.StatusInternalServerError)

		return
	}

	//TODO: Do
}

func (h *HttpHandler[K]) handleJoinRequest(responseWriter http.ResponseWriter, request *http.Request) {
	requestBody := map[string]string{}
	if decodeErr := json.NewDecoder(request.Body).Decode(&requestBody); decodeErr != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)

		return
	}

	if len(requestBody) != 2 {
		responseWriter.WriteHeader(http.StatusBadRequest)

		return
	}

	remoteAddress, ok := requestBody["address"]
	if !ok {
		responseWriter.WriteHeader(http.StatusBadRequest)

		return
	}

	nodeId, ok := requestBody["id"]
	if !ok {
		responseWriter.WriteHeader(http.StatusBadRequest)

		return
	}

	if joinErr := h.store.AddNode(remoteAddress, nodeId); joinErr != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)

		return
	}
}

func (h *HttpHandler[K]) handleUpdateLocationRequest(responseWriter http.ResponseWriter, request *http.Request) {
	if !h.EnableIngress {
		responseWriter.WriteHeader(http.StatusInternalServerError)

		return
	}

	//TODO: Do
}

func NewHttpHandler[K stores.StoreKey](store StoreNode[K]) *HttpHandler[K] {
	return &HttpHandler[K]{
		store: store,
	}
}

func (h *HttpHandler[K]) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		switch request.URL.Path {
		case "/location":
			h.handleGetLocationRequest(responseWriter, request)
		}
	case http.MethodPost:
		switch request.URL.Path {
		case "/join":
			h.handleJoinRequest(responseWriter, request)
		case "/location":
			h.handleUpdateLocationRequest(responseWriter, request)
		default:
			responseWriter.WriteHeader(http.StatusNotFound)
		}
	default:
		responseWriter.WriteHeader(http.StatusNotFound)
	}
}
