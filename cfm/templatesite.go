package cfm

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/coderconvoy/dbase"
	"github.com/coderconvoy/lazyf"
	"github.com/coderconvoy/templater/tempower"
)

//templateSite (*) implements ConfigItem and can be used by the muxer, once it has chosen the appropriate domain.
type templateSite struct {
	//The host which will redirect to the folder
	DomList
	//The Folder containing this hosting, -multiple hosts may point to the same folder
	Folder string
	//The Filename inside the Folder of the file we watch for changes
	Modifier string
	lastMod  time.Time
	plates   *tempower.PowerTemplate
	sync.Mutex
}

type Loose struct {
	FileName string
	Style    string
}

func NewTemplateSite(lz lazyf.LZ, root string) (*templateSite, error) {
	fol, err := lz.PString("folder", "Folder")
	if err != nil || len(fol) == 0 {
		return nil, errors.New("No folder for item")
	}
	if fol[0] != '/' {
		fol = path.Join(root, fol)
	}

	item := &templateSite{
		DomList:  lz.PStringAr("host", "Host"),
		Folder:   fol,
		Modifier: lz.PStringD("modify", "Mod", "mod", "Modifier", "modifier"),
	}
	item.lastMod = item.mod()

	plates, err := tempower.NewPowerTemplate(path.Join(fol, "templates/*"), fol)
	if err != nil {
		return nil, err
	}
	item.plates = plates

	return item, nil
}

func (ts *templateSite) Plates() *tempower.PowerTemplate {
	ts.Lock()
	res := ts.plates
	ts.Unlock()
	return res
}

func (ts *templateSite) Update() {
	nts := ts.mod()
	if nts.Equal(ts.lastMod) {
		return
	}

	newPlates, err := tempower.NewPowerTemplate(path.Join(ts.Folder, "templates/*"), ts.Folder)
	if err != nil {
		ts.Log("Could not Update templates:" + err.Error())
		return
	}

	ts.Lock()
	ts.plates = newPlates
	ts.Unlock()
	ts.lastMod = nts
}

func (ts *templateSite) mod() time.Time {
	mpath := path.Join(ts.Folder, ts.Modifier)
	t, err := GetModified(mpath)
	if err != nil {
		return time.Time{}
	}
	return t
}

func (ts *templateSite) Log(s string) {
	go func() {
		ts.Lock()
		defer ts.Unlock()
		logToFolder(path.Join(ts.Folder, "logs"), s)

	}()
}

//TryTemplate is the main useful method takes
//w: io writer
//host: the request.URL.Host
//p:the template name
//data:The data to send to the template
func (ts *templateSite) TryTemplate(w io.Writer, host string, p string, data interface{}) error {
	t := ts.Plates()

	b := new(bytes.Buffer)
	err := t.ExecuteTemplate(b, p, data)
	if err != nil {
		dbase.QLog(err.Error())
		return err
	}
	io.Copy(w, b)
	return nil
}

func (ts *templateSite) GetFilePath(fname string) (string, error) {
	rpath := ts.Folder
	if len(rpath) == 0 {
		return "", errors.New("No Folder location for host:")
	}

	res, err := SafeJoin(rpath, fname)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (ts *templateSite) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//Static files under /s/

	if strings.HasPrefix(r.URL.Path, "/s/") {
		fPath, err := ts.GetFilePath(r.URL.Path)
		if err != nil {
			w.Write([]byte("Bad File Request"))
		}
		http.ServeFile(w, r, fPath)
		return
	}

	//Handle restyling options with a style cookie
	host := r.Host
	styleC, cerr := r.Cookie("style")
	style := ""
	if cerr == nil {
		style = styleC.Value
	}

	s2 := r.URL.Query().Get("style")
	if s2 != "" {
		style = s2
		styleC = &http.Cookie{Name: "style", Value: style, Expires: time.Now().Add(time.Hour * 24)}
	}

	if styleC != nil {
		http.SetCookie(w, styleC)
	}

	//allow errors

	var err error
	p := strings.TrimPrefix(r.URL.Path, "/")

	LogTof(host, "Path---%s", p)
	// Empty for index

	if p == "" {
		err = ts.TryTemplate(w, host, "index", Loose{"index.md", style})
		if err != nil {
			fmt.Fprintf(w, "Could not load index, err = %s", err)
			dbase.QLogf("Could not load index, err = %s", err)
		}
		return
	}

	//Top Level fake to s
	fp, err := ts.GetFilePath("s/" + p)
	if err == nil {
		_, err2 := os.Stat(fp)
		if err2 == nil {
			http.ServeFile(w, r, fp)
			return
		}
	}
	//try template
	errlist := []error{}

	for k, v := range p {
		if v == '/' {
			err = ts.TryTemplate(w, host, p[:k], Loose{p[k+1:], style})
			if err != nil {
				dbase.QLog("No template found:" + err.Error())
			}

			if err == nil {
				return
			}
			errlist = append(errlist, err)
			dbase.QLog(err)
		}
	}
	err = ts.TryTemplate(w, host, p, Loose{"", style})
	if err == nil {
		return
	}
	errlist = append(errlist, err)

	err = ts.TryTemplate(w, host, "loose", Loose{p + ".md", style})
	if err != nil {
		dbase.QLog(err)
		errlist = append(errlist, err)
		for _, e := range errlist {
			fmt.Fprintln(w, e)
		}
	}

}
