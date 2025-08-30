package raft

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ristryder/maydinhed/mapping"
)

type HttpHandler[K mapping.ClusteredMarkerKey] struct {
	markerCluster     *mapping.MarkerCluster[K]
	nodeConfiguration *NodeConfiguration
	store             StoreNode[K]
}

func (h *HttpHandler[K]) handleGetLocationRequest(responseWriter http.ResponseWriter, request *http.Request) {
	if !h.nodeConfiguration.EnableEgress {
		responseWriter.WriteHeader(http.StatusInternalServerError)

		return
	}

	//TODO: Do
	locations, locationsErr := h.markerCluster.GetLocations(mapping.BoundingBox{}, 1)
	if locationsErr != nil {
		log.Println(locationsErr)

		responseWriter.WriteHeader(http.StatusInternalServerError)

		return
	}

	locationsBytes, marshalErr := json.Marshal(locations)
	if marshalErr != nil {
		log.Println(marshalErr)

		responseWriter.WriteHeader(http.StatusInternalServerError)

		return
	}

	responseWriter.WriteHeader(http.StatusOK)

	responseWriter.Write(locationsBytes)
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

	responseWriter.WriteHeader(http.StatusOK)
}

func (h *HttpHandler[K]) handleUpdateLocationRequest(responseWriter http.ResponseWriter, request *http.Request) {
	if !h.nodeConfiguration.EnableIngress {
		responseWriter.WriteHeader(http.StatusInternalServerError)

		return
	}

	updateLocationRequest := &mapping.Location{}
	if decodeErr := json.NewDecoder(request.Body).Decode(&updateLocationRequest); decodeErr != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)

		return
	}

	//TODO: Do
	var TODO_KEY interface{} = "test-key"

	if !h.nodeConfiguration.ValidateLocationUpdate {
		h.store.SetAndForget(TODO_KEY.(K), *updateLocationRequest)

		responseWriter.WriteHeader(http.StatusOK)

		return
	}

	if setResult := h.store.Set(TODO_KEY.(K), *updateLocationRequest); setResult != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)

		return
	}

	responseWriter.WriteHeader(http.StatusOK)
}

func NewHttpHandler[K mapping.ClusteredMarkerKey](markerCluster mapping.MarkerCluster[K], nodeConfiguration NodeConfiguration, store StoreNode[K]) *HttpHandler[K] {
	return &HttpHandler[K]{
		markerCluster:     &markerCluster,
		nodeConfiguration: &nodeConfiguration,
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
