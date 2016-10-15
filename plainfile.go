package templater

import (
	"bytes"
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
var globRoots map[string]string = make(map[string]string, 0)

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

func AddGlobRoot(k, v string) {
	globRoots[k] = v
}

//GetDirList is an attempt at safeguarding a file look up for templates.
//I want the programmer to be able to define where templates may look for files.
func GetDirList(rootK, dname string) ([]os.FileInfo, error) {
	if len(globRoots) == 0 {
		return []os.FileInfo{}, errors.New("No Safe directories set for GetDirList")
	}

	rooty, ok := globRoots[rootK]

	if !ok {
		return []os.FileInfo{}, errors.New("No valid folder")
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

func ParseHeadedMD(b []byte) map[string]string {
	ss := []string{"\n#\n", "\n\r#\n\r", "\r#\r"}
	splitP := -1
	l := -1
	for _, v := range ss {
		splitP = bytes.Index(b, []byte(v))
		if splitP >= 0 {
			l = len(v)
			break
		}
	}
	if splitP == -1 {
		return map[string]string{"contents": string(b)}

	}

	//TODO make this separate the bits
	return map[string]string{
		"contents": string(blackfriday.MarkdownCommon(b[splitP+l:])),
		"head":     string(b[:splitP]),
	}
}

//GetSharedHeadedMD returns first a map with heads and contents separated out
//Most importantly processed markdown will be in the map as contents
//other map elements should include tags,css,style
func GetSharedHeadedMD(fname string) map[string]string {
	return ParseHeadedMD(GetSharedFile(fname))
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
