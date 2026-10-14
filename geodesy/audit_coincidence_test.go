package geodesy_test

import (
	"math"
	"testing"

	geo "github.com/faustbrian/go-geo"
	"github.com/faustbrian/go-geo/geodesy"
)

func TestGeographicCoincidenceHasNoDefinedBearing(t *testing.T) {
	for _, model := range []geodesy.Model{geodesy.MeanEarthSphere(), geodesy.WGS84Ellipsoid()} {
		for _, pair := range []struct {
			name string
			from [2]float64
			to   [2]float64
		}{
			{name: "north pole", from: [2]float64{0, 90}, to: [2]float64{90, 90}},
			{name: "south pole", from: [2]float64{-180, -90}, to: [2]float64{45, -90}},
			{name: "dateline equator", from: [2]float64{-180, 0}, to: [2]float64{180, 0}},
			{name: "dateline latitude", from: [2]float64{-180, 37}, to: [2]float64{180, 37}},
		} {
			t.Run(model.Name()+"/"+pair.name, func(t *testing.T) {
				from := coordinate(t, pair.from[0], pair.from[1], geo.WGS84())
				to := coordinate(t, pair.to[0], pair.to[1], geo.WGS84())
				for _, direction := range [][2]geo.Coordinate{{from, to}, {to, from}} {
					result, err := model.Inverse(direction[0], direction[1])
					if err != nil {
						t.Fatal(err)
					}
					if result.BearingsDefined() || result.Distance().Meters() != 0 {
						t.Errorf("same geographic location: distance=%g, bearings defined=%v", result.Distance().Meters(), result.BearingsDefined())
					}
				}
			})
		}
		// Coincidence detection is exact geography, not an epsilon that erases
		// a short but real path or ordinary bearings.
		result, err := model.Inverse(coordinate(t, 0, 0, geo.WGS84()), coordinate(t, 0.00000001, 0, geo.WGS84()))
		if err != nil || !result.BearingsDefined() || result.Distance().Meters() <= 0 {
			t.Fatalf("%s collapsed distinct nearby locations: distance=%g, defined=%v, err=%v", model.Name(), result.Distance().Meters(), result.BearingsDefined(), err)
		}
	}
}

func TestAdjacentDatelineCoordinatesRemainDistinct(t *testing.T) {
	for _, model := range []geodesy.Model{geodesy.MeanEarthSphere(), geodesy.WGS84Ellipsoid()} {
		for _, latitude := range []float64{0, 37} {
			for _, longitude := range []float64{-180, 180} {
				from := coordinate(t, longitude, latitude, geo.WGS84())
				to := coordinate(t, math.Nextafter(-longitude, 0), latitude, geo.WGS84())
				for _, direction := range [][2]geo.Coordinate{{from, to}, {to, from}} {
					result, err := model.Inverse(direction[0], direction[1])
					if err != nil || !result.BearingsDefined() || result.Distance().Meters() <= 0 {
						t.Errorf("%s collapsed distinct adjacent dateline locations %v -> %v: distance=%g, defined=%v, err=%v", model.Name(), direction[0], direction[1], result.Distance().Meters(), result.BearingsDefined(), err)
					}
				}
			}
		}
	}
}
