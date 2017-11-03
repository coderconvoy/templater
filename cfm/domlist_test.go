package cfm

import (
	"testing"
)

func Test_DomList(t *testing.T) {
	d := DomList{"storyfeet.com", "www.happy"}

	ts := []struct {
		d string
		r bool
	}{
		{"www.storyfeet.com", true},
		{"www.storyfeet", false},
		{"com", false},
		{"storyfeet.com", true},
		{"happy", false},
	}

	for k, v := range ts {
		res := d.CanHost(v.d)
		if res != v.r {
			t.Errorf("%d:%s: expected %t, got %t", k, v.d, v.r, res)
		}
	}
}
