package cfm

import (
	"bytes"
	"errors"
	"io"
	"path"
	"sync"
	"time"

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
		return err
	}
	io.Copy(w, b)
	return nil
}

func (ts templateSite) GetFilePath(fname string) (string, error) {
	rpath := ts.Folder
	if len(rpath) == 0 {
		return "", errors.New("No Folder location for host:")
	}
	if ts.Folder[0] != '/' {
		rpath = path.Join(man.rootLoc, rpath)
	}

	res, err := SafeJoin(rpath, fname)
	if err != nil {
		return "", err
	}
	return res, nil
}
