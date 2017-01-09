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
