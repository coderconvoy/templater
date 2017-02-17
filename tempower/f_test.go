package tempower

import (
	"github.com/coderconvoy/templater/parse"
	"io/ioutil"
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

func Test_getFile(t *testing.T) {
	fmap := fileGetter("test_data/getfile/")

	getHMD := fmap["headedMDFile"].(func(string) (map[string]string, error))

	hmd, err := getHMD("popple.md")
	if err != nil {
		t.Log("popple not loaded", err)
		t.Fail()
	}

	popf, _ := ioutil.ReadFile("test_data/getfile/popple.md")
	popmap := parse.Headed(popf)
	popmap["md"] = mdParse(popmap["contents"])

	for k, v := range popmap {
		if v != hmd[k] {
			t.Logf("%s, not equal\nexpected: %s\ngot: %s\n", k, v, hmd[k])
			t.Fail()
		}
	}

	_, err = getHMD("p/../../poop.html")
	if err == nil {
		t.Logf("Reached out of locked directory for poop.html")
		t.Fail()
	}

}

func Test_filtercontains(t *testing.T) {
	td := []string{
		"hello",
		"loeh",
		"kol",
		"lop",
	}
	loRes := filterContains(td, "lo")
	loExp := []string{
		"hello",
		"loeh",
		"lop",
	}
	if len(loRes) != len(loExp) {
		t.Logf("lo - expected len %d, got len %d\n", len(loExp), len(loRes))
		t.FailNow()
	}
	for k, v := range loExp {
		if loRes[k] != v {
			t.Logf("lo - expected %s, got %s\n", v, loRes[k])
			t.Fail()
		}
	}

}

func Test_multireplace(t *testing.T) {
	td := []string{
		"hello",
		"loeh",
		"kol",
		"lop",
	}
	loRes := multiReplace(td, "lo", "kk", -1)
	loExp := []string{
		"helkk",
		"kkeh",
		"kol",
		"kkp",
	}
	if len(loRes) != len(loExp) {
		t.Logf("lo - expected len %d, got len %d\n", len(loExp), len(loRes))
		t.FailNow()
	}
	for k, v := range loExp {
		if loRes[k] != v {
			t.Logf("lo - expected %s, got %s\n", v, loRes[k])
			t.Fail()
		}
	}

}
