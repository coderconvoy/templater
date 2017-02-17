package blob

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func Test_Loader(t *testing.T) {
	bs := NewBlobSet("test_data")

	recs, err := bs.GetDir("s", "")
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
	fm, killer := SafeBlobFuncs("test_data")

	pinf, err := fm["getblobdir"].(func(string, ...string) ([]PageInfo, error))("s")
	if err != nil {
		t.Logf("Chan Err:%s", err)
		t.FailNow()
	}
	if len(pinf) != 3 {
		t.Logf("wrong len: expected 3 , got %d", len(pinf))
		t.Fail()
	}

	getblob := fm["getblob"].(func(string, string, ...string) (map[string]string, error))
	mp, err := getblob("s", "purple.md")
	if err != nil {
		t.Log("getblob1 error back")
		t.FailNow()
	}
	if mp["title"] != "My Favourite Color" {
		t.Logf("title expected 'My Favourite Color' got %s", mp["title"])
		t.Logf(mp["contents"])
		t.Fail()
	}
	killer()

	_, err = getblob("s", "purple.md")
	if err == nil {
		t.Log("No Error on closed chan fail")
		t.Fail()
	}

}

func test_ReadDir(t *testing.T) {
	d, _ := ioutil.ReadDir("test_data")
	for _, v := range d {
		fmt.Println(v.Name())
	}
}

func Test_GetNames(t *testing.T) {
	fm, _ := SafeBlobFuncs("test_data")

	f := fm["getblobnames"].(func(string, ...string) ([]string, error))

	r, err := f("s")
	if err != nil {
		t.Logf("Error on read s, %s\n", err)
		t.FailNow()
	}
	if len(r) != 3 {
		t.Logf("Expected 3 members, got %d", len(r))
		t.FailNow()
	}
	for _, v := range r {
		fmt.Println(v)
	}

}
