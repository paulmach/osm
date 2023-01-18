package annotate

import (
	"errors"
	"testing"

	"github.com/onXmaps/osm/annotate/internal/core"
)

func TestMapErrors(t *testing.T) {
	e := mapErrors(&core.NoHistoryError{})
	if _, ok := e.(*NoHistoryError); !ok {
		t.Errorf("should map NoHistoryError: %+v", e)
	}

	e = mapErrors(&core.NoVisibleChildError{})
	if _, ok := e.(*NoVisibleChildError); !ok {
		t.Errorf("should map NoVisibleChildError: %+v", e)
	}

	err := errors.New("some error")
	if e := mapErrors(err); e != err {
		t.Errorf("should pass through other errors: %v", e)
	}
}
