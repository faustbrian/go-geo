package geogeom_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/twpayne/go-geom"

	geo "github.com/faustbrian/go-geo"
	legacy "github.com/faustbrian/go-geo/adapter/gogeom"
	geogeom "github.com/faustbrian/go-geo/adapters/geom"
)

func TestFromGoGeomRejectsUnrepresentableSRID(t *testing.T) {
	if strconv.IntSize < 64 {
		t.Skip("out-of-range positive SRIDs require a 64-bit input int")
	}
	for _, srid := range []int64{1<<32 + 4326, -(1 << 32) + 4326} {
		for name, convert := range map[string]func(geom.T, geo.Limits) (geo.Geometry, error){
			"canonical": geogeom.FromGoGeom,
			"legacy":    legacy.FromGoGeom, //nolint:staticcheck // Verify the retained facade's public contract.
		} {
			t.Run(name+"/"+strconv.FormatInt(srid, 10), func(t *testing.T) {
				point := geom.NewPointFlat(geom.XY, []float64{24, 60}).SetSRID(int(srid))
				converted, err := convert(point, geo.DefaultLimits())
				var typed *geo.CRSError
				if converted != nil || !errors.As(err, &typed) {
					t.Fatalf("FromGoGeom out-of-range SRID: result=%v error=%v, want nil and CRSError", converted, err)
				}
				if typed.SRID != 0 {
					t.Fatalf("unrepresentable SRID reported as %d, want unset", typed.SRID)
				}
			})
		}
	}
}

func TestFromGoGeomPreservesRepresentableSRID(t *testing.T) {
	for _, srid := range []int{1, 1<<31 - 1} {
		point := geom.NewPointFlat(geom.XY, []float64{24, 60}).SetSRID(srid)
		converted, err := geogeom.FromGoGeom(point, geo.DefaultLimits())
		if err != nil || converted.CRS().SRID() != int32(srid) {
			t.Fatalf("FromGoGeom SRID %d: result=%v error=%v", srid, converted, err)
		}
	}
}

func TestFromGoGeomRejectsUnrepresentableDescendantSRID(t *testing.T) {
	if strconv.IntSize < 64 {
		t.Skip("out-of-range positive SRIDs require a 64-bit input int")
	}
	for _, srid := range []int64{1<<32 + 4326, -(1 << 32) + 4326} {
		for name, convert := range map[string]func(geom.T, geo.Limits) (geo.Geometry, error){
			"canonical": geogeom.FromGoGeom,
			"legacy":    legacy.FromGoGeom, //nolint:staticcheck // Verify the retained facade's public contract.
		} {
			for _, nested := range []bool{false, true} {
				t.Run(name+"/"+strconv.FormatInt(srid, 10)+"/nested="+strconv.FormatBool(nested), func(t *testing.T) {
					var child geom.T = geom.NewPointFlat(geom.XY, []float64{24, 60}).SetSRID(int(srid))
					if nested {
						child = geom.NewGeometryCollection().MustPush(geom.NewPointFlat(geom.XY, []float64{24, 60})).SetSRID(int(srid))
					}
					outer := geom.NewGeometryCollection().MustPush(child).SetSRID(4326)
					converted, err := convert(outer, geo.DefaultLimits())
					var typed *geo.CRSError
					if converted != nil || !errors.As(err, &typed) || typed.SRID != 0 {
						t.Fatalf("descendant SRID admission: result=%v error=%v, want nil and unrepresentable CRSError", converted, err)
					}
				})
			}
		}
	}
}

func TestFromGoGeomPreservesInheritedDescendantSRID(t *testing.T) {
	child := geom.NewPointFlat(geom.XY, []float64{24, 60})
	outer := geom.NewGeometryCollection().MustPush(child).SetSRID(4326)
	converted, err := geogeom.FromGoGeom(outer, geo.DefaultLimits())
	if err != nil || converted == nil || converted.CRS().SRID() != 4326 {
		t.Fatalf("inherited child SRID: result=%v error=%v", converted, err)
	}
}
