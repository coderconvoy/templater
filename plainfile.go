package templater

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

var initialised bool = false
var paths []string = make([]string, 0)

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

func GetSharedFile(libname string) []byte {
	setup()
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
