//configmanager is a holder for all the separate hosts and the folders they represent.
//A configuration reads a json file containing an array [] of ConfigItem s
//If the last 'Host' is 'default' this will be a catch all
package cfm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/coderconvoy/dbase"
	"github.com/coderconvoy/templater/blob"
	"github.com/coderconvoy/templater/tempower"
	"github.com/coderconvoy/templater/timestamp"
	"github.com/pkg/errors"

	"io"
	"io/ioutil"
	"sync"
	"time"
)

type temroot struct {
	root      string
	modifier  string
	templates *tempower.PowerTemplate
	last      time.Time
}

type ConfigItem struct {
	//The host which will redirect to the folder
	Host string
	//The Folder containing this hosting, -multiple hosts may point to the same folder
	Folder string
	//The Filename inside the Folder of the file we watch for changes
	Modifier string
}

type Manager struct {
	filename string
	tmap     map[string]*temroot
	config   []ConfigItem
	killflag bool
	sync.Mutex
}

//NewManager Creates a new Manager from json file based on ConfigItem
//params cFileName the name of the file
func NewManager(cFileName string) (*Manager, error) {
	c, err := loadConfig(cFileName)
	if err != nil {
		return nil, err
	}

	temps := newTMap(c)

	res := &Manager{
		filename: cFileName,
		config:   c,
		tmap:     temps,
		killflag: false,
	}

	go manageTemplates(res)

	return res, nil
}

//TryTemplate is the main useful method takes
//w: io writer
//host: the request.URL.Host
//p:the template name
//data:The data to send to the template
func (man *Manager) TryTemplate(w io.Writer, host string, p string, data interface{}) error {

	b := new(bytes.Buffer)
	for i := 0; i < 10; i++ {
		t, err := man.getTemplates(host)
		if err != nil {
			return err
		}
		err = t.ExecuteTemplate(b, p, data)
		if err == nil {
			w.Write(b.Bytes())
			return nil
		}
		if err != blob.DeadBlob() {
			return err
		}
	}

	return fmt.Errorf("Tried too many times to access blob")
}

func (man *Manager) GetConfig(host string) (ConfigItem, error) {

	host = strings.TrimPrefix(host, "www.")
	for _, v := range man.config {
		if v.Host == host || v.Host == "default" {
			return v, nil
		}
	}
	return ConfigItem{}, errors.New("No config assigned to that name")

}

func (man *Manager) GetFilePath(host, fname string) (string, error) {
	c, err := man.GetConfig(host)
	if err != nil {
		return "", errors.Wrap(err, "Could not get file path")
	}

	res := path.Join(c.Folder, fname)
	if !strings.HasPrefix(res, c.Folder) {
		return "", errors.New("No Upwards path building")
	}
	return res, nil

}

//Kill ends all internal go routines. Do not use the manager after calling Kill()
func (man *Manager) Kill() {
	man.Lock()
	defer man.Unlock()

	man.killflag = true
	//TODO loop through and kill all templates
	for _, v := range man.tmap {
		v.templates.Kill()
	}
}

func newTemroot(fol, mod string) (temroot, error) {
	tpath := path.Join(fol, "templates/*.*")
	dbase.QLogf("New Path = %s", tpath)
	t, err := tempower.NewPowerTemplate(tpath, fol)
	if err != nil {
		return temroot{}, err
	}
	return temroot{
		root:      fol,
		modifier:  mod,
		templates: t,
		last:      time.Now(),
	}, nil

}

func newTMap(conf []ConfigItem) map[string]*temroot {
	res := make(map[string]*temroot)

	for _, v := range conf {
		_, ok := res[v.Folder]
		if !ok {
			t, err := newTemroot(v.Folder, v.Modifier)
			if err == nil {
				res[v.Folder] = &t
			} else {
				dbase.QLogf("Could not load templates :%s,%s", v.Folder, err)
			}
		}
	}
	return res
}

func loadConfig(fName string) ([]ConfigItem, error) {
	var configs []ConfigItem

	b, err := ioutil.ReadFile(fName)
	if err != nil {
		return configs, err
	}

	err = json.Unmarshal(b, &configs)
	if err != nil {
		return configs, err
	}
	return configs, nil
}

func manageTemplates(man *Manager) {

	lastCheck := time.Now()
	var thisCheck time.Time

	for {
		thisCheck = time.Now()

		//if config has been updated then reset everything
		ts, err := timestamp.GetMod(man.filename)
		if err == nil {

			if ts.After(lastCheck) {
				dbase.QLog("Config File Changed")
				newcon, err := loadConfig(man.filename)
				if err == nil {
					oldmap := man.tmap
					man.Lock()
					man.config = newcon
					man.tmap = newTMap(man.config)
					man.Unlock()

					for _, v := range oldmap {
						v.templates.Kill()
					}

				} else {
					//ignore the change
					dbase.QLog("Load Config Error:", err)
				}
			}
		}

		//check folders for update only update the changed
		for k, v := range man.tmap {
			modpath := path.Join(v.root, v.modifier)
			ts, err := timestamp.GetMod(modpath)
			if err == nil {
				if ts.After(v.last) {
					t, err2 := newTemroot(v.root, v.modifier)
					if err2 == nil {
						man.Lock()
						man.tmap[k] = &t
						v.templates.Kill()
						v.last = ts
						man.Unlock()
					} else {
						dbase.QLogf("ERROR , Could not parse templates Using old ones: %s,%s\n", modpath, err2)
					}

				}

			} else {
				dbase.QLogf("ERROR, Mod file missing:%s,%s\n ", modpath, err)
			}

		}

		//Allow kill
		if man.killflag {
			return
		}
		//for each file look at modified file if changed update.
		lastCheck = thisCheck
		time.Sleep(time.Minute / 2)
	}

}

func (man *Manager) getTemplates(host string) (*tempower.PowerTemplate, error) {
	c, err := man.GetConfig(host)
	if err != nil {
		return nil, errors.Wrap(err, "No config available for host: "+host)
	}

	man.Lock()
	defer man.Unlock()
	t, ok := man.tmap[c.Folder]

	if !ok {
		return nil, errors.New("No Templates available for host: " + host)
	}
	return t.templates, nil

}
