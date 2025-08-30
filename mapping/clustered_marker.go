package mapping

type ClusteredMarker[K ClusteredMarkerKey] struct {
	IsClustered bool

	clusterArea   *BoundingBox
	keyedLocation *KeyedLocation[K]
	pixelX        int64
	pixelY        int64
}

type ClusteredMarkerKey interface {
	comparable
}

func NewClusteredMarker[K ClusteredMarkerKey](keyedLocation KeyedLocation[K]) *ClusteredMarker[K] {
	return &ClusteredMarker[K]{
		clusterArea:   NewDefaultBoundingBox(),
		keyedLocation: &keyedLocation,
		pixelX:        -1,
		pixelY:        -1,
	}
}

func NewDefaultClusteredMarker[K ClusteredMarkerKey]() *ClusteredMarker[K] {
	return &ClusteredMarker[K]{
		clusterArea: NewDefaultBoundingBox(),
		pixelX:      -1,
		pixelY:      -1,
	}
}

func (c *ClusteredMarker[K]) AddMarker(clusteredMarker ClusteredMarker[K]) {
	if c.keyedLocation == nil {
		c.keyedLocation = clusteredMarker.keyedLocation
	}

	c.clusterArea.Include(*clusteredMarker.clusterArea)
}

func (c *ClusteredMarker[K]) PixelX(zoomLevel int) int64 {
	if c.pixelX < 0 {
		c.pixelX = ConvertLongitudeToPixelX(c.keyedLocation.Longitude, zoomLevel)
	}

	return c.pixelX
}

func (c *ClusteredMarker[K]) PixelY(zoomLevel int) int64 {
	if c.pixelY < 0 {
		c.pixelY = ConvertLatitudeToPixelY(c.keyedLocation.Latitude, zoomLevel)
	}

	return c.pixelY
}
