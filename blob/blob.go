package blob

import (
	"fmt"
	"github.com/coderconvoy/templater/parse"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"
	"time"
)

type PageInfo struct {
	Title string
	FName string
	Date  time.Time
}

type BlobSet map[string][]PageInfo

type ByDateDown []PageInfo

func (bd ByDateDown) Len() int           { return len(bd) }
func (bd ByDateDown) Less(a, b int) bool { return bd[b].Date.Before(bd[a].Date) }
func (bd ByDateDown) Swap(a, b int)      { bd[a], bd[b] = bd[b], bd[a] }

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
		res = append(res, PageInfo{t, v.Name(), dt})

		f.Close()
	}

	sort.Sort(ByDateDown(res))

	(*bs)[fol] = res

	return res, nil
}

func (bs *BlobSet) GetBlob(fol, file string) map[string]string {
	infos, err := bs.GetDir(fol)
	if err != nil {
		return map[string]string{}
	}

	for _, v := range infos {
		if v.FName == file || v.Title == file {

			f, err := ioutil.ReadFile(path.Join(fol, v.FName))
			if err != nil {
				return map[string]string{
					"title":    "File Not Loaded",
					"contents": err.Error(),
				}

			}
			return parse.HeadedMD(f)

		}
	}

	return map[string]string{
		"title":    "Not Found",
		"contents": fmt.Sprintf("Could not find \"%s\" in \"%s\"", file, fol),
	}

}

func BlobGetter() chan func(BlobSet) {
	ch := make(chan func(BlobSet))

	go func() {
		bb := BlobSet{}
		for fn := range ch {
			fn(bb)
		}
	}()
	return ch
}

func AccessMap(ch chan func(BlobSet)) template.FuncMap {
	type backinfo struct {
		pi  []PageInfo
		err error
	}
	getAll := func(fol string) ([]PageInfo, error) {
		bchan := make(chan backinfo)
		ch <- func(bs BlobSet) {
			bi, er := bs.GetDir(fol)
			res := backinfo{bi, er}
			bchan <- res
		}
		res := <-bchan
		return res.pi, res.err
	}

	getOne := func(fol, file string) map[string]string {
		bchan := make(chan map[string]string)
		ch <- func(bs BlobSet) {
			bchan <- bs.GetBlob(fol, file)
		}
		return <-bchan

	}

	return template.FuncMap{
		"getblobdir": getAll,
		"getblob":    getOne,
	}
}
