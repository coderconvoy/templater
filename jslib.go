package templater

import (
	"fmt"
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
	jspaths := os.Getenv("GO_JSPATH")

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
	paths = append(paths, newPath)
}

func GetLib(libname string) []byte {
	setup()
	for i := len(paths) - 1; i >= 0; i-- {
		res, err := ioutil.ReadFile(path.Join(paths[i], libname))
		if err == nil {
			return res
		}
	}
	return make([]byte, 0)
}

func ServeLib(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/js/")

	path2 := strings.Replace(path, "../", "", -1)
	for path2 != path {
		path = path2
		path2 = strings.Replace(path, "../", "", -1)
	}
	fmt.Printf("Path = :%s\n\n", path)

	w.Write(GetLib(path))
}
