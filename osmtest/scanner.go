package osmtest

import "github.com/paulmach/osm"

// Scanner implements the osm.Scanner interface with
// just a list of elements.
type Scanner struct {
	// ScanError can be used to trigger an error.
	// If non-nil, Next() will return false and Err() will
	// return this error.
	ScanError error

	offset   int
	elements osm.Elements
}

var _ osm.Scanner = &Scanner{}

// NewScanner creates a new test scanner useful for test stubbing.
func NewScanner(elements osm.Elements) *Scanner {
	return &Scanner{
		offset:   -1,
		elements: elements,
	}
}

// Scan progresses the scanner to the next element.
func (s *Scanner) Scan() bool {
	if s.ScanError != nil {
		return false
	}

	s.offset++
	return s.offset < len(s.elements)
}

// Element returns the current element.
func (s *Scanner) Element() osm.Element {
	return s.elements[s.offset]
}

// Err returns the scanner.ScanError.
func (s *Scanner) Err() error {
	return s.ScanError
}

// Close is a stub for this test scanner.
func (s *Scanner) Close() error {
	return nil
}
