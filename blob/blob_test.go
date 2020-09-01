package blob

import (
	"testing"
)

func compDir(t *testing.T, dt []PageInfo, exp []string) {
	if len(dt) != len(exp) {
		t.Errorf("Not same num elems :%d , %d", len(dt), len(exp))
	}

	for k, v := range dt {
		if v.FName != exp[k] {
			t.Errorf("Diff elems")
		}
	}
}

func Test_Loader(t *testing.T) {
	bs := NewBlobSet("test_data")

	recs, err := bs.GetDir("s", "chapter")
	if err != nil {
		t.Errorf("Get Error")
	}
	compDir(t, recs, []string{"purple.md", "t2.md", "t1.md"})

	recs, err = bs.GetDir("s", "-chapter")
	if err != nil {
		t.Errorf("Get Error")
	}
	compDir(t, recs, []string{"t1.md", "t2.md", "purple.md"})

}
