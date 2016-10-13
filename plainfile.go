package templater

import (
	"errors"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

var initialised bool = false
var paths []string = make([]string, 0)
var globRoots []string = make([]string, 0)

func setup() {
	if initialised {
		return
	}
	initialised = true
	jspaths := os.Getenv("GO_SHARE")

	if jspaths == "" {
		return
	}
	ar := strings.Split(jspaths, ":")
	for _, v := range ar {
		if v != "" {
			paths = append(paths, ar...)
		}
	}

}

func AddPath(newPath string) {
	setup()
	paths = append(paths, newPath)
}

func AddGlobRoot(newPath string) {
	globRoots = append(globRoots, newPath)
}

//GetDirList is an attempt at safeguarding a file look up for templates.
//I want the programmer to be able to define where templates may look for files.
func GetDirList(dname string, root ...string) ([]os.FileInfo, error) {
	if len(globRoots) == 0 {
		return []os.FileInfo{}, errors.New("No Safe directories set for GetDirList")
	}
	rooty := globRoots[0]
	if len(root) > 0 {
		for _, v := range globRoots {
			if root[0] == v {
				rooty = v
			}
		}
	}
	if strings.Index(dname, "../") >= 0 {
		return []os.FileInfo{}, errors.New("No upward paths allowed")
	}

	return ioutil.ReadDir(path.Join(rooty, dname))

}

func GetSharedFile(libname string) []byte {
	setup()
	if strings.Index(libname, "../") >= 0 {
		return []byte("No Upward paths (\"../\" allowed")
	}
	for i := len(paths) - 1; i >= 0; i-- {
		res, err := ioutil.ReadFile(path.Join(paths[i], libname))
		if err == nil {
			return res
		}
	}
	return make([]byte, 0)
}

func GetSharedLines(libname string) []string {
	s := string(GetSharedFile(libname))
	res := strings.Split(s, "\n")
	return res
}

//GetSharedMD will return a parsed MD File from a lib
func GetSharedMD(fname string) string {
	return string(blackfriday.MarkdownCommon(GetSharedFile(fname)))
}

//ServeSharedFile will serve the file from any of the shared paths it looks in, prefering those added later.
//Any paths including "../" are refused
func ServeSharedFile(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/share/")

	if strings.Index(path, "../") >= 0 {
		w.Write([]byte("No Upward paths (\"../\" allowed"))
		return
	}

	w.Write(GetSharedFile(path))
}
