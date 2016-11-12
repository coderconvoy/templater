package blob

import (
	"fmt"
	"github.com/coderconvoy/templater/parse"
	"path"
)

type PageInfo struct {
	Title string
	Path  string
	Date  string
	Head  string
}

type Blobset map[string][]PageInfo

type ByDate []PageInfo

func (bd ByDate) Len()          { return len(bd) }
func (bd ByDate) Less(a, b int) { return bd[a].Date < bd[b].Date }
func (bd ByDate) Swap(a, b int) { bd[a], bd[b] = bd[b], bd[a] }

func (bs *Blobset) GetDir(fol string) []PageInfo {
	if res, ok := bs[fol]; ok {
		return res
	}

	d, err := os.ReadDir(fol)
	if err != nil {
		fmt.Println("Dir Not Available")
		return nil, fmt.Errorf("GetDir could not read dir %s", err)
	}
	res = make([]PageInfo, 0)

	for i, v := range d {
		fPath = path.Join(fol, v.Name())
		f, err = os.Open(fPath)
		if err != nil {
			continue
		}
		//TODO get data our of file and into res

		f.Close()
	}

	return res
}

func BlobGetter() chan func(bs Blobset) {
	ch := make(chan func(bs Blobset))

	go func() {
		bb := Blobset{}
		for fn := range ch {
			fn(bs)
		}
	}()
	return ch

}
