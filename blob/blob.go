//The Purpose of the blob package is to provide a quick way of grabbing headed files from a template
//calling blob.SafeBlobFuncs(string) returns a map of functions all wrapped in clojures to access the same BlobSet Blobs the header data from each blob is stored in ram, for quick listing and archiving in the main use
package blob

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"text/template"
	"time"

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
	*sync.Mutex
}

func NewBlobSet(root string) *BlobSet {
	return &BlobSet{root, make(map[string][]PageInfo), &sync.Mutex{}}
}

func (bs *BlobSet) GetDir(fol string, sortBy string) ([]PageInfo, error) {
	if sortBy != "name" {
		sortBy = "date"
	}

	fol = path.Join(bs.root, fol)
	store := sortBy + "#" + fol

	bs.Lock()
	defer bs.Unlock()
	if res, ok := bs.m[store]; ok {
		return res, nil
	}

	d, err := ioutil.ReadDir(fol)
	if err != nil {
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
		Sort(res, ByName(true))
	} else {
		Sort(res, ByDate(false))
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

func AccessMap(rootFol string) template.FuncMap {
	cache := NewBlobSet(rootFol)

	getAll := func(fol string, sortMode ...string) ([]PageInfo, error) {
		sm := "date"
		if len(sortMode) > 0 {
			sm = sortMode[0]
		}
		res, err := cache.GetDir(fol, sm)

		if err != nil {
			return nil, err
		}
		return res, err
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

	getOne := func(fol, file string, sortMode ...string) map[string]string {
		sm := "date"
		if len(sortMode) > 0 {
			sm = sortMode[0]
		}
		return cache.GetBlob(fol, file, sm)
	}

	getOneMD := func(fol, file string, sortMode ...string) (map[string]string, error) {
		res := getOne(fol, file, sortMode...)
		res["md"] = string(blackfriday.MarkdownCommon([]byte(res["contents"])))
		res["contents"] = res["md"] // Covering legacy,
		return res, nil

	}

	getContent := func(fol, file string, sortMode ...string) (string, error) {
		res := getOne(fol, file, sortMode...)

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
