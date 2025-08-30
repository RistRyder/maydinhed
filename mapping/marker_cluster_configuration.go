package mapping

import "time"

type MarkerClusterConfiguration struct {
	autoClusterInterval time.Duration
}

func NewMarkerClusterConfiguration(autoClusterInterval time.Duration) *MarkerClusterConfiguration {
	return &MarkerClusterConfiguration{
		autoClusterInterval: autoClusterInterval,
	}
}
