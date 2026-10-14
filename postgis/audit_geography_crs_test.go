package postgis_test

import (
	"errors"
	"testing"

	geo "github.com/faustbrian/go-geo"
	"github.com/faustbrian/go-geo/postgis"
)

func TestGeographyPredicateRequiresExplicitWGS84Metadata(t *testing.T) {
	column, err := postgis.NewColumn("locations.coordinate")
	if err != nil {
		t.Fatal(err)
	}
	distance, _ := geo.NewDistanceMeters(10)
	for _, test := range []struct {
		name      string
		srid      int32
		label     string
		wantError bool
	}{
		{name: "canonical control", srid: 4326, label: "EPSG:4326"},
		{name: "mismatched label", srid: 4326, label: "caller-label", wantError: true},
		{name: "foreign SRID control", srid: 3857, label: "EPSG:3857", wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			crs, err := geo.NewCRS(test.srid, test.label)
			if err != nil {
				t.Fatal(err)
			}
			lon, _ := geo.NewLongitude(1)
			lat, _ := geo.NewLatitude(2)
			coordinate, err := geo.NewCoordinate(lon, lat, crs)
			if err != nil {
				t.Fatal(err)
			}
			point, err := geo.NewPoint(coordinate)
			if err != nil {
				t.Fatal(err)
			}
			fragment, err := postgis.GeographyDWithin(column, point, distance, 3)
			if test.wantError {
				if !errors.Is(err, geo.ErrCRS) || fragment.SQL() != "" || len(fragment.Args()) != 0 {
					t.Fatalf("foreign metadata reached SQL: %q, %v", fragment.SQL(), err)
				}
				return
			}
			if err != nil || fragment.SQL() != `ST_DWithin("locations"."coordinate"::geography, $3::geography, $4)` || len(fragment.Args()) != 2 || fragment.Args()[1] != float64(10) {
				t.Fatalf("canonical bound SQL changed: %q, %v, %v", fragment.SQL(), fragment.Args(), err)
			}
			value, ok := fragment.Args()[0].(postgis.Value)
			if !ok {
				t.Fatalf("geometry argument lost its codec owner: %T", fragment.Args()[0])
			}
			owned, valid := value.Geometry()
			if !valid || !geo.EqualGeometry(owned, point) {
				t.Fatal("bound geography argument changed geometry or CRS")
			}
			encoded, err := value.Value()
			if err != nil || encoded == nil {
				t.Fatalf("bound geography argument cannot be encoded: %v", err)
			}
		})
	}
}
