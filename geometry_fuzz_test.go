package geo_test

import (
	"testing"

	geo "github.com/faustbrian/go-geo"
)

func FuzzGeometryConstructors(f *testing.F) {
	f.Add([]byte{0, 0, 127, 127, 255, 255})
	f.Add([]byte{128, 64, 32, 16})

	f.Fuzz(func(_ *testing.T, data []byte) {
		if len(data) > 256 {
			data = data[:256]
		}
		coordinates := make([]geo.Coordinate, 0, len(data)/2+1)
		for index := 0; index+1 < len(data); index += 2 {
			longitude, _ := geo.NewLongitude(float64(data[index])*360/255 - 180)
			latitude, _ := geo.NewLatitude(float64(data[index+1])*180/255 - 90)
			coordinate, _ := geo.NewCoordinate(longitude, latitude, geo.WGS84())
			coordinates = append(coordinates, coordinate)
		}
		if len(coordinates) == 0 {
			longitude, _ := geo.NewLongitude(0)
			latitude, _ := geo.NewLatitude(0)
			coordinate, _ := geo.NewCoordinate(longitude, latitude, geo.WGS84())
			coordinates = append(coordinates, coordinate)
		}

		point, _ := geo.NewPoint(coordinates[0])
		line, lineErr := geo.NewLineString(coordinates)
		ring := append(append([]geo.Coordinate(nil), coordinates...), coordinates[0])
		polygon, polygonErr := geo.NewPolygon(ring, nil)
		_, _ = geo.NewMultiPoint(coordinates, geo.WGS84())
		if lineErr == nil {
			_, _ = geo.NewMultiLineString([]geo.LineString{line}, geo.WGS84())
		}
		if polygonErr == nil {
			_, _ = geo.NewMultiPolygon([]geo.Polygon{polygon}, geo.WGS84())
		}
		_, _ = geo.NewGeometryCollection([]geo.Geometry{point}, geo.WGS84())
	})
}
