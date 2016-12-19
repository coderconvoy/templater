package tempower

import (
	"testing"
)

func Test_templater1(t *testing.T) {
	_, _ = NewPowerTemplate("test_data/*.html", "test_data")
}

func Test_getN(t *testing.T) {
	d := []int{4, 5, 8, 3, 6, 9, 34, 12, 34}
	c, _ := getN(4, d)
	cd := c.([]int)
	if len(cd) != 4 {
		t.Logf("Len c == %d, expected 4.\n", len(cd))
		t.Fail()
	}
}
