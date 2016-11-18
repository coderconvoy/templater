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

type ClosedSetError string

func (cse ClosedSetError) Error() string {
	return string(cse)
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
		if v.FName == file || v.Title == file || file == "" {

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

func SafeBlobFuncs() (template.FuncMap, func()) {
	ch, saf := blobChans()

	runner, killer := blobGetter(ch, saf)

	return AccessMap(runner), killer
}

func blobGetter(ch chan func(BlobSet), safety chan bool) (func(func(BlobSet)) error, func()) {

	runner := func(f func(BlobSet)) error {
		if <-safety {
			ch <- f
			return nil
		}
		return ClosedSetError("Closed Set Blob")

	}

	killer := func() {
		if <-safety {
			close(ch)
		}
	}

	return runner, killer
}

func blobChans() (chan func(BlobSet), chan bool) {
	ch := make(chan func(BlobSet))
	safety := make(chan bool)

	go func() {
		bb := BlobSet{}
		safety <- true
		for fn := range ch {
			fn(bb)
			safety <- true
		}
		close(safety)
	}()
	return ch, safety

}

func AccessMap(runner func(func(BlobSet)) error) template.FuncMap {
	type backinfo struct {
		pi  []PageInfo
		err error
	}

	getAll := func(fol string) ([]PageInfo, error) {
		bchan := make(chan backinfo)
		err := runner(func(bs BlobSet) {
			bi, er := bs.GetDir(fol)
			res := backinfo{bi, er}
			bchan <- res
		})

		if err != nil {
			return nil, err
		}
		res := <-bchan
		return res.pi, res.err
	}

	getOne := func(fol, file string) (map[string]string, error) {
		bchan := make(chan map[string]string)
		err := runner(func(bs BlobSet) {
			bchan <- bs.GetBlob(fol, file)
		})
		if err != nil {
			return nil, err
		}
		return <-bchan, nil

	}

	return template.FuncMap{
		"getblobdir": getAll,
		"getblob":    getOne,
	}
}
