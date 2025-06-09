package tours

import (
	"testing"
	"time"
)

func TestDateUnmarshalText(t *testing.T) {
	input := "2025-03-30"
	var d Date
	err := d.UnmarshalText([]byte(input))
	if err != nil {
		t.Error(err)
	}

	if d != (Date{2025, time.March, 30}) {
		t.Errorf("unexpected date: %v", d)
	}
}
