package editor

import (
	"testing"
)

func Test_Build(t *testing.T) {
	e := NewEditor("test_data")

	_, err := e.ListR("../", 2)
	if err == nil {
		t.Log("Able to go up a dir")
		t.FailNow()
	}

	b, err := e.ListR("popdir", 1)
	if err != nil {
		t.Log("could not ListR 'popdir'", err)
		t.FailNow()
	}

	if len(b) != 1 {
		t.Log("more than one file in popdir")
		t.FailNow()
	}

	b, err = e.ListR("", 2)
	if err != nil {
		t.Log("Could not ListR '' in test_data", err)
		t.FailNow()
	}

	dirFound := false

	for _, v := range b {
		if v.Info.IsDir() {
			dirFound = true
			if v.Children == nil {
				t.Log("directory with no children")
				t.Fail()
			}
		}
	}

	if !dirFound {
		t.Log("No Dir found, not even popdir")
		t.Fail()
	}

}
