package mapping

import (
	"context"
	"log"
	"maps"
	"math"
	"slices"
	"time"
)

const (
	clusterHeight = 27
	clusterWidth  = 22
)

type MarkerCluster[K ClusteredMarkerKey] struct {
	autoClusteringCtx          context.Context
	autoClusteringCtxCancel    context.CancelFunc
	clusteredMarkers           map[K]ClusteredMarker[K]
	markerClusterConfiguration *MarkerClusterConfiguration
}

func addMarkersInRange[K ClusteredMarkerKey](destinationClusteredMarker *ClusteredMarker[K], clusteredMarkers []ClusteredMarker[K], index, direction, zoomLevel int) {
	clusteredMarkerCount := len(clusteredMarkers)
	searchIndex := index + direction

	for {
		if searchIndex >= clusteredMarkerCount || searchIndex < 0 {
			return
		}

		if !clusteredMarkers[searchIndex].IsClustered {
			if math.Abs(float64(clusteredMarkers[searchIndex].PixelX(zoomLevel)-clusteredMarkers[index].PixelX(zoomLevel))) < clusterWidth {
				if math.Abs(float64(clusteredMarkers[searchIndex].PixelY(zoomLevel)-clusteredMarkers[index].PixelY(zoomLevel))) < clusterHeight {
					destinationClusteredMarker.AddMarker(clusteredMarkers[searchIndex])

					clusteredMarkers[searchIndex].IsClustered = true
				}
			} else {
				return
			}
		}

		searchIndex += direction
	}
}

func (m *MarkerCluster[K]) cluster(clusteredMarkers []ClusteredMarker[K], zoomLevel int) error {
	slices.SortStableFunc(clusteredMarkers, func(a, b ClusteredMarker[K]) int {
		if a.keyedLocation.Longitude > b.keyedLocation.Longitude {
			return 1
		} else {
			if a.keyedLocation.Longitude == b.keyedLocation.Longitude {
				if a.keyedLocation.Latitude > a.keyedLocation.Latitude {
					return 1
				} else {
					if a.keyedLocation.Latitude == b.keyedLocation.Latitude {
						return 0
					} else {
						return -1
					}
				}
			} else {
				return -1
			}
		}
	})

	newlyClusteredMarkers := []ClusteredMarker[K]{}

	for i, clusteredMarker := range clusteredMarkers {
		if clusteredMarker.IsClustered {
			continue
		}

		currentClusteredMarker := NewDefaultClusteredMarker[K]()
		currentClusteredMarker.AddMarker(clusteredMarker)
		currentClusteredMarker.IsClustered = true

		addMarkersInRange(currentClusteredMarker, clusteredMarkers, i, -1, zoomLevel)

		addMarkersInRange(currentClusteredMarker, clusteredMarkers, i, 1, zoomLevel)

		newlyClusteredMarkers = append(newlyClusteredMarkers, *currentClusteredMarker)
	}

	return nil
}

func (m *MarkerCluster[K]) startAutoClusteringLoop() {
	log.Println("starting marker auto clustering")

	go func() {
		for {
			select {
			case <-m.autoClusteringCtx.Done():
				log.Println("stopping marker auto clustering")
				return
			case <-time.After(m.markerClusterConfiguration.autoClusterInterval):
				clusteredMarkers := slices.Collect(maps.Values(m.clusteredMarkers))
				//TODO: Do
				m.cluster(clusteredMarkers, 2)
			}
		}
	}()
}

func NewMarkerCluster[K ClusteredMarkerKey](markerClusterConfiguration MarkerClusterConfiguration) *MarkerCluster[K] {
	return &MarkerCluster[K]{
		clusteredMarkers:           make(map[K]ClusteredMarker[K]),
		markerClusterConfiguration: &markerClusterConfiguration,
	}
}

func (m *MarkerCluster[K]) Add(clusteredMarker ClusteredMarker[K]) error {
	m.clusteredMarkers[clusteredMarker.keyedLocation.Id] = clusteredMarker

	return nil
}

func (m *MarkerCluster[K]) GetLocations(boundingBox BoundingBox, zoomLevel int) ([]Location, error) {
	//TODO: Do
	locations := []Location{}
	for clusteredMarker := range maps.Values(m.clusteredMarkers) {
		locations = append(locations, clusteredMarker.keyedLocation.Location)
	}

	return locations, nil
}

func (m *MarkerCluster[K]) StartAutoClustering() error {
	if m.autoClusteringCtxCancel != nil {
		return nil
	}

	m.autoClusteringCtx, m.autoClusteringCtxCancel = context.WithCancel(context.Background())

	m.startAutoClusteringLoop()

	return nil
}

func (m *MarkerCluster[K]) StopAutoClustering() error {
	if m.autoClusteringCtxCancel == nil {
		return nil
	}

	m.autoClusteringCtxCancel()

	m.autoClusteringCtxCancel = nil

	return nil
}
