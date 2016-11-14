package blob

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func Test_Loader(t *testing.T) {
	bs := BlobSet{}

	recs, err := bs.GetDir("test_data")
	if err != nil {
		t.Log("Get Error")
		t.FailNow()
	}
	if len(recs) != 3 {
		t.Logf("Not enough infos expected 3 , got : %d\n", len(recs))
		t.FailNow()
	}

	for _, v := range recs {
		fmt.Println(v)
	}

}

func TestChannelAccess(t *testing.T) {
	fm := AccessMap(BlobGetter())

	pinf, err := fm["getblobdir"].(func(string) ([]PageInfo, error))("test_data")
	if err != nil {
		t.Logf("Chan Err:%s", err)
		t.FailNow()
	}
	if len(pinf) != 3 {
		t.Logf("wrong len: expected 3 , got %d", len(pinf))
		t.Fail()
	}

	mp := fm["getblob"].(func(string, string) map[string]string)("test_data", "purple.md")
	if mp["title"] != "My Favourite Color" {
		t.Logf("title expected 'My Favourite Color' got %s", mp["title"])
		t.Logf(mp["contents"])
		t.Fail()
	}
}

func test_ReadDir(t *testing.T) {
	d, _ := ioutil.ReadDir("test_data")
	for _, v := range d {
		fmt.Println(v.Name())
	}
}
