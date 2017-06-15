//The Purpose of the blob package is to provide a quick way of grabbing headed files from a template
//calling blob.SafeBlobFuncs(string) returns a map of functions all wrapped in clojures to access the same BlobSet Blobs the header data from each blob is stored in ram, for quick listing and archiving in the main use
package blob

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/coderconvoy/dbase"
	"github.com/coderconvoy/templater/parse"
	"github.com/russross/blackfriday"
)

type PageInfo struct {
	Title string
	FName string
	Date  time.Time
}

type BlobSet struct {
	root string
	m    map[string][]PageInfo
}

func NewBlobSet(root string) *BlobSet {
	return &BlobSet{root, make(map[string][]PageInfo)}
}

type DeadBlobErr struct{}

func (dbe DeadBlobErr) Error() string {
	return "This blob has been killed"
}

var deadBlob = DeadBlobErr{}

func DeadBlob() error {
	return deadBlob
}

type ByDateDown []PageInfo
type ByName []PageInfo

func (bd ByDateDown) Len() int           { return len(bd) }
func (bd ByDateDown) Less(a, b int) bool { return bd[b].Date.Before(bd[a].Date) }
func (bd ByDateDown) Swap(a, b int)      { bd[a], bd[b] = bd[b], bd[a] }

func (bd ByName) Len() int           { return len(bd) }
func (bd ByName) Less(a, b int) bool { return bd[b].FName > bd[a].FName }
func (bd ByName) Swap(a, b int)      { bd[a], bd[b] = bd[b], bd[a] }

func (bs *BlobSet) GetDir(fol string, sortBy string) ([]PageInfo, error) {
	if sortBy != "name" {
		sortBy = "date"
	}

	fol = path.Join(bs.root, fol)
	store := sortBy + "#" + fol
	if res, ok := bs.m[store]; ok {
		return res, nil
	}

	d, err := ioutil.ReadDir(fol)
	if err != nil {
		dbase.QLogf("Dir Not Available %s", fol)
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

	if sortBy == "name" {
		sort.Sort(ByName(res))
	} else {
		sort.Sort(ByDateDown(res))
	}

	bs.m[store] = res

	return res, nil
}

func (bs *BlobSet) GetBlob(fol, file, sortMode string) map[string]string {
	infos, err := bs.GetDir(fol, sortMode)
	if err != nil {
		return map[string]string{}
	}
	file = strings.ToLower(file)

	for i, v := range infos {
		if strings.ToLower(v.FName) == file || strings.ToLower(v.Title) == file || file == "" {

			f, err := ioutil.ReadFile(path.Join(bs.root, fol, v.FName))
			if err != nil {
				return map[string]string{
					"title":    "File Not Loaded",
					"contents": err.Error(),
					"FName":    v.FName,
				}

			}
			res := parse.Headed(f)
			res["FName"] = v.FName
			if i > 0 {
				res["next"] = infos[i-1].FName
			}
			if i+1 < len(infos) {
				res["prev"] = infos[i+1].FName
			}
			return res

		}
	}

	return map[string]string{
		"title":    "Not Found",
		"contents": fmt.Sprintf("Could not find \"%s\" in \"%s\"", file, fol),
	}

}

func SafeBlobFuncs(root string) (template.FuncMap, func()) {
	ch, saf := blobChans(root)

	runner, killer := blobGetter(ch, saf)

	return AccessMap(runner), killer
}

func blobGetter(ch chan func(*BlobSet), safety chan bool) (func(func(*BlobSet)) error, func()) {

	runner := func(f func(*BlobSet)) error {
		if <-safety {
			ch <- f
			return nil
		}
		return deadBlob
	}

	killer := func() {
		if <-safety {
			close(ch)
		}
	}

	return runner, killer
}

func blobChans(root string) (chan func(*BlobSet), chan bool) {
	ch := make(chan func(*BlobSet))
	safety := make(chan bool)

	go func() {
		bb := NewBlobSet(root)
		safety <- true
		for fn := range ch {
			fn(bb)
			safety <- true
		}
		close(safety)
	}()
	return ch, safety

}

func AccessMap(runner func(func(*BlobSet)) error) template.FuncMap {
	type backinfo struct {
		pi  []PageInfo
		err error
	}

	getAll := func(fol string, sortMode ...string) ([]PageInfo, error) {
		sm := "date"
		if len(sortMode) > 0 {
			sm = sortMode[0]
		}

		bchan := make(chan backinfo)
		err := runner(func(bs *BlobSet) {
			bi, er := bs.GetDir(fol, sm)
			res := backinfo{bi, er}
			bchan <- res
		})

		if err != nil {
			return nil, err
		}
		res := <-bchan
		return res.pi, res.err
	}

	getAllNames := func(fol string, sortMode ...string) ([]string, error) {
		a, err := getAll(fol, sortMode...)
		if err != nil {
			return []string{}, err
		}
		res := make([]string, len(a))
		for k, v := range a {
			res[k] = v.FName
		}
		return res, nil

	}

	getOne := func(fol, file string, sortMode ...string) (map[string]string, error) {
		sm := "date"
		if len(sortMode) > 0 {
			sm = sortMode[0]
		}
		bchan := make(chan map[string]string)
		err := runner(func(bs *BlobSet) {
			bchan <- bs.GetBlob(fol, file, sm)
		})
		if err != nil {
			return nil, err
		}
		return <-bchan, nil

	}

	getOneMD := func(fol, file string, sortMode ...string) (map[string]string, error) {
		res, err := getOne(fol, file, sortMode...)
		if err != nil {
			return nil, err
		}
		res["contents"] = string(blackfriday.MarkdownCommon([]byte(res["contents"])))
		return res, nil

	}

	getContent := func(fol, file string, sortMode ...string) (string, error) {
		res, err := getOne(fol, file, sortMode...)
		if err != nil {
			return "", err
		}

		return res["contents"], nil
	}

	return template.FuncMap{
		"getblobdir":   getAll,
		"getblobnames": getAllNames,
		"getblob":      getOne,
		"getblobMD":    getOneMD,
		"getblobc":     getContent,
	}
}
