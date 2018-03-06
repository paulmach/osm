package osmapi

import (
	"errors"
	"testing"
)

func TestDatasourceNotFound(t *testing.T) {
	ds := NewDatasource(nil)

	if ds.NotFound(nil) {
		t.Errorf("should be false for nil")
	}

	if ds.NotFound(errors.New("foo")) {
		t.Errorf("should be false for random error")
	}

	if ds.NotFound(&GoneError{}) {
		t.Errorf("should be false for gone error")
	}

	if !ds.NotFound(&NotFoundError{}) {
		t.Errorf("should be true for not found error")
	}
}
