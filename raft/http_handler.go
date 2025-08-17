package raft

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/govalues/decimal"
	"github.com/ristryder/maydinhed/stores"
)

type HttpHandler[K stores.StoreKey] struct {
	nodeConfiguration NodeConfiguration
	store             StoreNode[K]
}

type updateLocationRequest[K stores.StoreKey] struct {
	Id       K               `json:"id"`
	Location stores.Location `json:"l"`
}

func (h *HttpHandler[K]) handleGetLocationRequest(responseWriter http.ResponseWriter, request *http.Request) {
	if !h.nodeConfiguration.EnableEgress {
		responseWriter.WriteHeader(http.StatusInternalServerError)

		return
	}

	//TODO: Do
	var key interface{} = "test-key"
	setErr := h.store.Set(key.(K), stores.Location{
		Latitude:  decimal.Hundred,
		Longitude: decimal.Hundred,
	})
	if setErr != nil {
		log.Println(setErr)
	}
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
	if !h.nodeConfiguration.EnableIngress {
		responseWriter.WriteHeader(http.StatusInternalServerError)

		return
	}

	updateLocationRequest := updateLocationRequest[K]{}
	if decodeErr := json.NewDecoder(request.Body).Decode(&updateLocationRequest); decodeErr != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)

		return
	}

	if !h.nodeConfiguration.ValidateLocationUpdate {
		h.store.SetAndForget(updateLocationRequest.Id, updateLocationRequest.Location)

		responseWriter.WriteHeader(http.StatusOK)

		return
	}

	if setResult := h.store.Set(updateLocationRequest.Id, updateLocationRequest.Location); setResult != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)

		return
	}

	responseWriter.WriteHeader(http.StatusOK)
}

func NewHttpHandler[K stores.StoreKey](nodeConfiguration NodeConfiguration, store StoreNode[K]) *HttpHandler[K] {
	return &HttpHandler[K]{
		nodeConfiguration: nodeConfiguration,
		store:             store,
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
