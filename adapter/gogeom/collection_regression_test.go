package gogeom_test

import (
	"fmt"
	"testing"

	"github.com/twpayne/go-geom"

	geo "github.com/faustbrian/go-geo"
	"github.com/faustbrian/go-geo/adapter/gogeom"
)

func TestReleasedGeometryCollectionPanicIsReplacedByOwnedConversion(t *testing.T) {
	collection := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{24, 60}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("FromGoGeom() panicked = %q, want owned conversion", fmt.Sprint(recovered))
		}
	}()
	if _, err := gogeom.FromGoGeom(collection, geo.DefaultLimits()); err != nil {
		t.Fatalf("FromGoGeom() error = %v", err)
	}
}
