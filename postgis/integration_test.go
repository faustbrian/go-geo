package postgis_test

import (
	"context"
	"math"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	geo "github.com/faustbrian/go-geo"
	"github.com/faustbrian/go-geo/geodesy"
	"github.com/faustbrian/go-geo/postgis"
)

func TestPostGISIntegration(t *testing.T) {
	dsn := os.Getenv("POSTGIS_DSN")
	if dsn == "" {
		t.Skip("POSTGIS_DSN is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	connection, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("pgx.Connect() error = %v", err)
	}
	defer func() {
		if closeErr := connection.Close(context.Background()); closeErr != nil {
			t.Errorf("Close() error = %v", closeErr)
		}
	}()

	var geometryOID, geographyOID uint32
	if err := connection.QueryRow(ctx, `
		SELECT 'geometry'::regtype::oid, 'geography'::regtype::oid
	`).Scan(&geometryOID, &geographyOID); err != nil {
		t.Fatalf("spatial OID query error = %v", err)
	}
	postgis.Register(connection.TypeMap(), geometryOID, geo.DefaultLimits())
	postgis.Register(connection.TypeMap(), geographyOID, geo.DefaultLimits())
	if _, err := connection.Exec(ctx, `
		CREATE TEMPORARY TABLE samples (
			id bigint PRIMARY KEY,
			geom geometry(Point, 4326) NOT NULL
		)
	`); err != nil {
		t.Fatalf("CREATE TABLE error = %v", err)
	}

	point := mustPoint(t, 24.9384, 60.1699)
	value, err := postgis.NewValue(point, geo.DefaultLimits())
	if err != nil {
		t.Fatalf("NewValue() error = %v", err)
	}
	if _, err := connection.Exec(ctx, "INSERT INTO samples (id, geom) VALUES ($1, $2)", 1, value); err != nil {
		t.Fatalf("INSERT error = %v", err)
	}
	var scanned postgis.Value
	if err := connection.QueryRow(ctx, "SELECT geom FROM samples WHERE id = $1", 1).Scan(&scanned); err != nil {
		t.Fatalf("SELECT geometry error = %v", err)
	}
	geometry, valid := scanned.Geometry()
	if !valid || !geo.EqualGeometry(geometry, point) {
		t.Fatal("PostGIS round trip changed geometry")
	}

	comparePostGISDistances(t, ctx, connection, point.Coordinate())
	executePostGISFragments(t, ctx, connection, point)
	comparePostGISPolygonLocations(t, ctx, connection)
}

func comparePostGISDistances(
	t *testing.T,
	ctx context.Context,
	connection *pgx.Conn,
	from geo.Coordinate,
) {
	t.Helper()
	to := mustCoordinate(t, -73.9857, 40.7484)
	var sphereDistance, spheroidDistance float64
	err := connection.QueryRow(ctx, `
		SELECT
			ST_DistanceSphere(
				ST_SetSRID(ST_MakePoint($1, $2), 4326),
				ST_SetSRID(ST_MakePoint($3, $4), 4326)
			),
			ST_DistanceSpheroid(
				ST_SetSRID(ST_MakePoint($1, $2), 4326),
				ST_SetSRID(ST_MakePoint($3, $4), 4326),
				'SPHEROID["WGS 84",6378137,298.257223563]'
			)
	`,
		from.Longitude().Degrees(),
		from.Latitude().Degrees(),
		to.Longitude().Degrees(),
		to.Latitude().Degrees(),
	).Scan(&sphereDistance, &spheroidDistance)
	if err != nil {
		t.Fatalf("PostGIS distance query error = %v", err)
	}
	spherical, err := geodesy.MeanEarthSphere().Inverse(from, to)
	if err != nil {
		t.Fatalf("spherical Inverse() error = %v", err)
	}
	ellipsoidal, err := geodesy.WGS84Ellipsoid().Inverse(from, to)
	if err != nil {
		t.Fatalf("ellipsoidal Inverse() error = %v", err)
	}
	// PostGIS rounds its mean-radius sphere differently from the explicit IUGG
	// 6,371,008.8 metre model, producing centimetre-scale long-haul deltas.
	if delta := math.Abs(sphereDistance - spherical.Distance().Meters()); delta > 0.05 {
		t.Fatalf("spherical distance differs from PostGIS by %v m", delta)
	}
	if delta := math.Abs(spheroidDistance - ellipsoidal.Distance().Meters()); delta > 0.001 {
		t.Fatalf("ellipsoidal distance differs from PostGIS by %v m", delta)
	}
}

func executePostGISFragments(
	t *testing.T,
	ctx context.Context,
	connection *pgx.Conn,
	point geo.Point,
) {
	t.Helper()
	column, err := postgis.NewColumn("samples.geom")
	if err != nil {
		t.Fatalf("NewColumn() error = %v", err)
	}
	distance, err := geo.NewDistanceMeters(1)
	if err != nil {
		t.Fatalf("NewDistanceMeters() error = %v", err)
	}
	dWithin, err := postgis.GeographyDWithin(column, point, distance, 1)
	if err != nil {
		t.Fatalf("GeographyDWithin() error = %v", err)
	}
	intersects, err := postgis.Intersects(column, point, 1)
	if err != nil {
		t.Fatalf("Intersects() error = %v", err)
	}
	for name, fragment := range map[string]postgis.Fragment{
		"dwithin":    dWithin,
		"intersects": intersects,
	} {
		var matched bool
		query := "SELECT " + fragment.SQL() + " FROM samples WHERE id = 1"
		if err := connection.QueryRow(ctx, query, fragment.Args()...).Scan(&matched); err != nil {
			t.Fatalf("%s query error = %v", name, err)
		}
		if !matched {
			t.Fatalf("%s predicate = false, want true", name)
		}
	}
}

func comparePostGISPolygonLocations(
	t *testing.T,
	ctx context.Context,
	connection *pgx.Conn,
) {
	t.Helper()
	polygon := mustPolygon(t,
		[][2]float64{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}},
		[][][2]float64{{{3, 3}, {3, 7}, {7, 7}, {7, 3}, {3, 3}}},
	)
	value, err := postgis.NewValue(polygon, geo.DefaultLimits())
	if err != nil {
		t.Fatalf("NewValue(polygon) error = %v", err)
	}
	probes := []struct {
		name      string
		longitude float64
		latitude  float64
	}{
		{name: "surface interior", longitude: 1, latitude: 1},
		{name: "shell boundary", longitude: 0, latitude: 5},
		{name: "hole interior", longitude: 5, latitude: 5},
		{name: "hole boundary", longitude: 3, latitude: 5},
		{name: "exterior", longitude: 11, latitude: 5},
	}
	for _, probe := range probes {
		coordinate := mustCoordinate(t, probe.longitude, probe.latitude)
		location, err := polygon.Locate(coordinate)
		if err != nil {
			t.Fatalf("Locate(%s) error = %v", probe.name, err)
		}
		var contains, touches bool
		err = connection.QueryRow(ctx, `
			SELECT
				ST_Contains($1::geometry, ST_SetSRID(ST_MakePoint($2, $3), 4326)),
				ST_Touches($1::geometry, ST_SetSRID(ST_MakePoint($2, $3), 4326))
		`, value, probe.longitude, probe.latitude).Scan(&contains, &touches)
		if err != nil {
			t.Fatalf("PostGIS location query (%s) error = %v", probe.name, err)
		}
		want := geo.Outside
		if touches {
			want = geo.Boundary
		} else if contains {
			want = geo.Inside
		}
		if location != want {
			t.Fatalf("Locate(%s) = %v, PostGIS = %v", probe.name, location, want)
		}
	}
}

func mustCoordinate(t *testing.T, longitude, latitude float64) geo.Coordinate {
	t.Helper()
	lon, err := geo.NewLongitude(longitude)
	if err != nil {
		t.Fatalf("NewLongitude() error = %v", err)
	}
	lat, err := geo.NewLatitude(latitude)
	if err != nil {
		t.Fatalf("NewLatitude() error = %v", err)
	}
	coordinate, err := geo.NewCoordinate(lon, lat, geo.WGS84())
	if err != nil {
		t.Fatalf("NewCoordinate() error = %v", err)
	}
	return coordinate
}

func mustPolygon(
	t *testing.T,
	exterior [][2]float64,
	holes [][][2]float64,
) geo.Polygon {
	t.Helper()
	convertRing := func(points [][2]float64) []geo.Coordinate {
		ring := make([]geo.Coordinate, len(points))
		for index, point := range points {
			ring[index] = mustCoordinate(t, point[0], point[1])
		}
		return ring
	}
	convertedHoles := make([][]geo.Coordinate, len(holes))
	for index, hole := range holes {
		convertedHoles[index] = convertRing(hole)
	}
	polygon, err := geo.NewPolygon(convertRing(exterior), convertedHoles)
	if err != nil {
		t.Fatalf("NewPolygon() error = %v", err)
	}
	return polygon
}
