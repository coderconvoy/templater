package blob

import (
	"fmt"
	"github.com/coderconvoy/templater/parse"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

type PageInfo struct {
	Title string
	FName string
	Date  time.Time
}

type BlobSet map[string][]PageInfo

type ByDate []PageInfo

func (bd ByDate) Len() int           { return len(bd) }
func (bd ByDate) Less(a, b int) bool { return bd[a].Date.Before(bd[b].Date) }
func (bd ByDate) Swap(a, b int)      { bd[a], bd[b] = bd[b], bd[a] }

func (bs *BlobSet) GetDir(fol string) ([]PageInfo, error) {
	if res, ok := (*bs)[fol]; ok {
		return res, nil
	}

	d, err := ioutil.ReadDir(fol)
	if err != nil {
		fmt.Println("Dir Not Available")
		return nil, fmt.Errorf("GetDir could not read dir %s", err)
	}
	res := make([]PageInfo, 0)

	for _, v := range d {
		if strings.HasPrefix(v.Name(), ".") {
			continue
		}
		fPath := path.Join(fol, v.Name())
		f, err := os.Open(fPath)
		if err != nil {
			continue
		}
		mp, _ := parse.HeadOnly(f)

		t, ok := mp["title"]
		if !ok {
			t = v.Name()
		}
		d, ok := mp["date"]
		dt := v.ModTime()
		if ok {
			pt, err := time.Parse("2/1/2006", d)
			if err != nil {
				pt, err = time.Parse("2/1/06", d)
			}

			if err == nil {
				dt = pt
			}
		}
		res = append(res, PageInfo{v.Name(), t, dt})

		f.Close()
	}

	sort.Sort(ByDate(res))

	(*bs)[fol] = res

	return res, nil
}

func BlobGetter() chan func(bs BlobSet) {
	ch := make(chan func(BlobSet))

	go func() {
		bb := BlobSet{}
		for fn := range ch {
			fn(bb)
		}
	}()
	return ch

}
