package configmanager

import (
	"os"
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

func Test_Manager(t *testing.T) {
	m, err := NewManager("test_data/test_load.json")
	if err != nil {
		t.Logf("Could not Load Manager :%s", err)
		t.FailNow()
	}

	err = m.TryTemplate(os.Stdout, "", "root", nil)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	m.Kill()

	err = m.TryTemplate(os.Stdout, "", "root", nil)
	if err == nil {
		t.Log("No Error using blob in dead template")
		t.Fail()
	}

}
