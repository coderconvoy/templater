package templater

import (
	"io/ioutil"
	"os"
)

func AgeSortedFiles(fpath string) []os.FileInfo {
	list := ioutil.ReadDir(fpath)

	return list
}
