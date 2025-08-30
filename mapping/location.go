package mapping

import (
	"math"
)

const (
	earthCircumference     = earthRadius * 2 * math.Pi
	earthHalfCircumference = earthCircumference / 2
	earthRadius            = 6378137
	pixelsPerTile          = 256
)

type KeyedLocation[K ClusteredMarkerKey] struct {
	Location
	Id K `json:"id"`
}

type Location struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
}

var InvalidLocation Location = Location{
	Latitude:  999,
	Longitude: 999,
}

func convertDegreesToRadians(degrees float64) float64 {
	return (degrees * math.Pi) / 180
}

func ConvertLatitudeToPixelY(latitude float64, zoomLevel int) int64 {
	shiftedZoomLevel := 1 << zoomLevel
	arc := earthCircumference / float64(shiftedZoomLevel*pixelsPerTile)
	latitudeRadians := convertDegreesToRadians(latitude)
	sinLatitude := math.Sin(latitudeRadians)
	metersY := earthRadius / 2 * math.Log((1+sinLatitude)/(1-sinLatitude))
	pixelY := math.Round((earthHalfCircumference - metersY) / arc)

	return int64(pixelY)
}

func ConvertLongitudeToPixelX(longitude float64, zoomLevel int) int64 {
	shiftedZoomLevel := 1 << zoomLevel
	arc := earthCircumference / float64(shiftedZoomLevel*pixelsPerTile)
	longitudeRadians := convertDegreesToRadians(longitude)
	metersX := earthRadius * longitudeRadians
	pixelX := math.Round((earthHalfCircumference + metersX) / arc)

	return int64(pixelX)
}

func NewLocation(latitude, longitude float64) *Location {
	return &Location{
		Latitude:  latitude,
		Longitude: longitude,
	}
}
