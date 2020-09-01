package cfm

import (
	"errors"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/coderconvoy/lazyf"
)

type staticSite struct {
	Folder    string
	LogFolder string
	PubFolder string
	DomList
	sync.Mutex
}

func NewStaticSite(lz lazyf.LZ, root string) (*staticSite, error) {
	fol, err := lz.PString("folder", "Folder")
	if err != nil || len(fol) == 0 {
		return nil, errors.New("No Root Folder for static site")
	}
	if fol[0] != '/' {
		fol = path.Join(root, fol)
	}
	println("static root fol = %s", fol)

	logFol, err := lz.PString("log", "Log")
	if err != nil || len(logFol) == 0 {
		logFol = "logs"
	}
	if logFol[0] != '/' {
		logFol = path.Join(fol, logFol)
	}
	//println("static logs = %s", logFol)
	pubFol, err := lz.PString("public", "Public")
	if err != nil || len(pubFol) == 0 {
		pubFol = "public"
	}
	if pubFol[0] != '/' {
		pubFol = path.Join(fol, pubFol)
	}
	//println("static pub = %s", pubFol)
	item := &staticSite{
		DomList:   lz.PStringAr("host", "Host"),
		Folder:    fol,
		LogFolder: logFol,
		PubFolder: pubFol,
	}
	return item, nil
}

//ConfigItem interface
func (ss *staticSite) Update() {
}

func (ss *staticSite) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//dbase.QLog(r.URL.Path)
	p := strings.TrimPrefix(r.URL.Path, "/")
	LogTof(r.Host, "Path---%s---remote(%s)", p, r.RemoteAddr)
	fPath, _ := ss.GetFilePath(r.URL.Path)
	http.ServeFile(w, r, fPath)
}

func (ss *staticSite) GetFilePath(fname string) (string, error) {

	rpath := ss.PubFolder
	if len(rpath) == 0 {
		return "", errors.New("No Folder location for host:")
	}

	safePath, err := SafeJoin(rpath, fname)
	if err != nil {
		return "", err
	}

	if fInfo, err := os.Stat(safePath); err == nil {
		if fInfo.IsDir() {
			return path.Join(safePath, "index.html"), nil
		}
	}

	return safePath, nil
}

func (ss *staticSite) Log(s string) {
	go func() {
		ss.Lock()
		defer ss.Unlock()
		logToFolder(ss.LogFolder, s)

	}()
}
