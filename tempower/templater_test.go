package tempower

import (
	"testing"
)

func Test_templater1(t *testing.T) {
	_ = NewPowerTemplate("test_data/*.html", "test_data")
}