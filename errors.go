package geo

import (
	"errors"
	"fmt"
)

var (
	// ErrRange classifies finite-value and numeric-range violations.
	ErrRange = errors.New("geo: value outside permitted range")
	// ErrTopology classifies structurally invalid geometry.
	ErrTopology = errors.New("geo: invalid topology")
	// ErrCRS classifies invalid or incompatible coordinate reference systems.
	ErrCRS = errors.New("geo: invalid coordinate reference system")
	// ErrEncoding classifies malformed or resource-exhausting encoded input.
	ErrEncoding = errors.New("geo: invalid encoding")
	// ErrUnsupported classifies well-formed operations the package cannot perform.
	ErrUnsupported = errors.New("geo: unsupported operation")
)

// RangeError reports a rejected numeric value and its inclusive bounds.
type RangeError struct {
	// ValueName names the scalar being validated.
	ValueName string
	// Value is the rejected value, possibly NaN or infinite.
	Value float64
	// Minimum is the inclusive lower admissible bound.
	Minimum float64
	// Maximum is the inclusive upper admissible bound.
	Maximum float64
}

// Error returns a diagnostic for the rejected value or operation.
func (e *RangeError) Error() string {
	return fmt.Sprintf(
		"geo: %s %v outside inclusive range [%v, %v]",
		e.ValueName,
		e.Value,
		e.Minimum,
		e.Maximum,
	)
}

// Unwrap returns the stable error category used by errors.Is.
func (e *RangeError) Unwrap() error { return ErrRange }

// TopologyError reports a geometry invariant violation.
type TopologyError struct {
	// Geometry names the affected geometry family or component.
	Geometry string
	// Problem describes the validation failure.
	Problem string
}

// Error returns a diagnostic for the rejected value or operation.
func (e *TopologyError) Error() string {
	return fmt.Sprintf("geo: invalid %s topology: %s", e.Geometry, e.Problem)
}

// Unwrap returns the stable error category used by errors.Is.
func (e *TopologyError) Unwrap() error { return ErrTopology }

// CRSError reports invalid or incompatible CRS metadata.
type CRSError struct {
	// SRID is the rejected or incompatible spatial reference identifier.
	SRID int32
	// Problem describes the validation failure.
	Problem string
}

// Error returns a diagnostic for the rejected value or operation.
func (e *CRSError) Error() string {
	return fmt.Sprintf("geo: invalid CRS SRID %d: %s", e.SRID, e.Problem)
}

// Unwrap returns the stable error category used by errors.Is.
func (e *CRSError) Unwrap() error { return ErrCRS }

// EncodingError reports malformed encoded input. Problem is a diagnostic;
// Cause may retain input from an underlying parser, so redact before logging.
type EncodingError struct {
	// Format names the codec or textual representation.
	Format string
	// Problem describes the validation failure.
	Problem string
	// Cause is an optional underlying error; it may retain caller input.
	Cause error
}

// Error returns a diagnostic for the rejected value or operation.
func (e *EncodingError) Error() string {
	return fmt.Sprintf("geo: invalid %s encoding: %s", e.Format, e.Problem)
}

// Unwrap returns the underlying Cause, or nil; use Is for ErrEncoding.
func (e *EncodingError) Unwrap() error { return e.Cause }

// Is classifies every EncodingError as ErrEncoding.
func (e *EncodingError) Is(target error) bool { return target == ErrEncoding }

// UnsupportedError reports a recognized operation that is not supported.
type UnsupportedError struct {
	// Operation names the unsupported operation.
	Operation string
	// Reason explains why the operation is unavailable.
	Reason string
}

// Error returns a diagnostic for the rejected value or operation.
func (e *UnsupportedError) Error() string {
	return fmt.Sprintf("geo: unsupported %s: %s", e.Operation, e.Reason)
}

// Unwrap returns the stable error category used by errors.Is.
func (e *UnsupportedError) Unwrap() error { return ErrUnsupported }
