package configmanager

import (
	"testing"
)

func Test_Create(t *testing.T) {
	c, err := loadConfig("test_data/nofile.json")
	if err == nil {
		t.Log("No file, but no error")
		t.Fail()
	}

	c, err = loadConfig("test_data/test_load.json")
	if err != nil {
		t.Logf("load test_load err = %s", err)
		t.Fail()
	}

	if len(c) != 2 {
		t.Logf("test_load expected 2 items, got %d", len(c))
		t.Fail()
	}

}
