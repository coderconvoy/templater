package cfm

import (
	"testing"
)

func Test_DomList(t *testing.T) {
	d := DomList{"www.storyfeet.com", "www.happy"}

	ts := []struct {
		d string
		r bool
	}{
		{"www.storyfeet", false},
		{"com", true},
		{"storyfeet.com", true},
		{"happy", true},
	}

	for k, v := range ts {
		res := d.CanHost(v.d)
		if res != v.r {
			t.Errorf("Line %d: expected %b, got %b", k, v.r, res)
		}
	}
}
