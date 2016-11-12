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

func Test_ReadDir(t *testing.T) {
	d, _ := ioutil.ReadDir("test_data")
	for _, v := range d {
		fmt.Println(v.Name())
	}
}
