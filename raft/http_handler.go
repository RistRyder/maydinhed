package raft

import (
	"encoding/json"
	"net/http"
)

type HttpHandler struct {
	store Store
}

func (h *HttpHandler) handleJoinRequest(responseWriter http.ResponseWriter, request *http.Request) {
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

func NewHttpHandler(store Store) *HttpHandler {
	return &HttpHandler{
		store: store,
	}
}

func (h *HttpHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.URL.Path == "/join" {
		h.handleJoinRequest(responseWriter, request)
	} else {
		responseWriter.WriteHeader(http.StatusNotFound)
	}
}
