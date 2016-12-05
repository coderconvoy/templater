//editor is the place where people can come and edit their files in a safe way.
//
package editor

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type Editor struct {
	Root string
}

type Branch struct {
	Info     os.FileInfo
	Children []*Branch
	IsLocked bool
}

func NewEditor(root string) *Editor {
	return &Editor{root}
}

func (ed *Editor) ListR(fpath string, depth int) ([]*Branch, error) {
	if depth <= 0 {
		return nil, nil
	}
	loc := path.Join(ed.Root, fpath)

	if !strings.HasPrefix(loc, ed.Root) {
		return nil, fmt.Errorf("New no upward paths allowed")
	}

	dr, err := ioutil.ReadDir(loc)
	if err != nil {
		return nil, err
	}

	res := make([]*Branch, len(dr))

	for k, v := range dr {

		var chids []*Branch
		if v.IsDir() {
			chids, err = ed.ListR(path.Join(fpath, v.Name()), depth-1)
		}
		//TODO set isLocked to the appropriate value
		res[k] = &Branch{Info: v, Children: chids, IsLocked: false}
	}
	return res, nil

}
