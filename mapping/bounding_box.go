package mapping

type BoundingBox struct {
	northWest *Location
	southEast *Location
}

func NewBoundingBox(northWest, southEast *Location) *BoundingBox {
	return &BoundingBox{
		northWest: northWest,
		southEast: southEast,
	}
}

func NewDefaultBoundingBox() *BoundingBox {
	return &BoundingBox{
		northWest: NewLocation(-90, 180),
		southEast: NewLocation(90, -180),
	}
}

func (b *BoundingBox) Include(boundingBox BoundingBox) {
	if boundingBox.southEast.Latitude < b.southEast.Latitude {
		b.southEast.Latitude = boundingBox.southEast.Latitude
	}
	if boundingBox.northWest.Latitude > b.northWest.Latitude {
		b.northWest.Latitude = boundingBox.northWest.Latitude
	}
	if boundingBox.southEast.Longitude > b.southEast.Longitude {
		b.southEast.Longitude = boundingBox.southEast.Longitude
	}
	if boundingBox.northWest.Longitude < b.northWest.Longitude {
		b.northWest.Longitude = boundingBox.northWest.Longitude
	}
}
