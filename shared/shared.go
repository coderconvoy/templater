//This package provided methods for the templater class to be used as an aid to building websites from template type files with minimal effort.
//Most of these functions can be added to a function map, however to give them access to their instance I have used clojures to lock the function to an instance of the Sharer struct
package shared

import (
	"errors"
	"github.com/coderconvoy/templater/parse"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

//Sharer is for keeping folder names, paths is places to look for files, globRoots are places to allow templates to search
type Sharer struct {
	paths     []string
	globRoots map[string]string
}

func NewSharer(useEnv bool, paths ...string) *Sharer {
	if useEnv {
		jspaths := os.Getenv("GO_SHARE")
		if jspaths != "" {
			paths = append(paths, strings.Split(jspaths, ":")...)

		}
	}
	return &Sharer{paths, make(map[string]string)}
}

func (self *Sharer) AddPath(newPath string) {
	self.paths = append(self.paths, newPath)
}

func (self *Sharer) AddSearchableRoot(k, v string) {
	self.globRoots[k] = v
}

//clojure to return the GetDirList function tied to an instance
func (self *Sharer) GetDirListF() func(string, string) ([]os.FileInfo, error) {
	return func(rootK, dname string) ([]os.FileInfo, error) {
		return self.GetDirList(rootK, dname)
	}
}

//GetDirList is an attempt at safeguarding a file look up for templates.
//I want the programmer to be able to define where templates may look for files.
func (self *Sharer) GetDirList(rootK, dname string) ([]os.FileInfo, error) {
	if len(self.globRoots) == 0 {
		return []os.FileInfo{}, errors.New("No Safe directories set for GetDirList")
	}

	rooty, ok := self.globRoots[rootK]

	if !ok {
		return []os.FileInfo{}, errors.New("No valid folder")
	}

	if strings.Index(dname, "../") >= 0 {
		return []os.FileInfo{}, errors.New("No upward paths allowed")
	}

	return ioutil.ReadDir(path.Join(rooty, dname))
}

//find and return file contents from possible paths
func (self *Sharer) GetFile(fpath string) []byte {
	if strings.Index(fpath, "../") >= 0 {
		return []byte("No Upward paths (\"../\" allowed")
	}
	for i := len(self.paths) - 1; i >= 0; i-- {
		res, err := ioutil.ReadFile(path.Join(self.paths[i], fpath))
		if err == nil {
			return res
		}
	}
	return make([]byte, 0)
}

//clojure for GetFile as Plain text
func (self *Sharer) GetFileTextF() func(string) string {
	return func(fpath string) string {
		return string(self.GetFile(fpath))
	}
}

//clojure for Get MD (Markdown) parsed file
func (self *Sharer) GetMDF() func(string) string {
	return func(s string) string {
		return self.GetMD(s)
	}
}

//GetMD will return a parsed MD File from a lib
func (self *Sharer) GetMD(fname string) string {
	return string(blackfriday.MarkdownCommon(self.GetFile(fname)))
}

//clojure for GetHeadedMDF
func (self *Sharer) GetHeadedMDF() func(string) map[string]string {
	return func(fname string) map[string]string {
		return self.GetHeadedMD(fname)
	}
}

//GetHeaded returns first a map with heads and contents separated out
//processed markdown will be in the map as contents
//other map elements should include tags,css,style
func (self *Sharer) GetHeadedMD(fname string) map[string]string {
	res := parse.Headed(self.GetFile(fname))
	res["contents"] = string(blackfriday.MarkdownCommon([]byte(res["contents"])))
	return res
}

//Get parser for MenuJSON
func (self *Sharer) GetJSONMenuF() func(string) (string, error) {
	return func(fname string) (string, error) {
		data := string(self.GetFile(fname))
		return parse.JSONMenu(data)
	}
}

func (self *Sharer) GetHTMLMenuF() func(string, string) (string, error) {
	return func(fname, rootID string) (string, error) {
		data := string(self.GetFile(fname))
		return parse.HTMLMenu(data, rootID)
	}
}

//Create a serverfunc that can read trim a prefix and return the appropriate shared file
func (self *Sharer) GetServerF(pfx ...string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		for _, v := range pfx {
			path = strings.TrimPrefix(path, v)
		}
		w.Write(self.GetFile(path))

	}
}
