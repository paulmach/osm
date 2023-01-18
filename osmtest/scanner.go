package osmtest

import "github.com/onXmaps/osm"

// Scanner implements the osm.Scanner interface with
// just a list of objects.
type Scanner struct {
	// ScanError can be used to trigger an error.
	// If non-nil, Next() will return false and Err() will
	// return this error.
	ScanError error

	offset  int
	objects osm.Objects
}

var _ osm.Scanner = &Scanner{}

// NewScanner creates a new test scanner useful for test stubbing.
func NewScanner(objects osm.Objects) *Scanner {
	return &Scanner{
		offset:  -1,
		objects: objects,
	}
}

// Scan progresses the scanner to the next object.
func (s *Scanner) Scan() bool {
	if s.ScanError != nil {
		return false
	}

	s.offset++
	return s.offset < len(s.objects)
}

// Object returns the current object.
func (s *Scanner) Object() osm.Object {
	return s.objects[s.offset]
}

// Err returns the scanner.ScanError.
func (s *Scanner) Err() error {
	return s.ScanError
}

// Close is a stub for this test scanner.
func (s *Scanner) Close() error {
	return nil
}
