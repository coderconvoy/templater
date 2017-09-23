//configmanager is a holder for all the separate hosts and the folders they represent.
//A configuration reads a json file containing an array [] of ConfigItem s
//If the last 'Host' is 'default' this will be a catch all
package cfm

import (
	"bytes"
	"encoding/json"
	"path"
	"strings"

	"github.com/coderconvoy/dbase"
	"github.com/coderconvoy/lazyf"
	"github.com/coderconvoy/templater/tempower"
	"github.com/pkg/errors"

	"io"
	"io/ioutil"
	"sync"
	"time"
)

type Manager struct {
	filename string
	rootLoc  string
	config   []ConfigItem
	lastEdit time.Time
	sync.Mutex
}

//NewManager Creates a new Manager from json file based on ConfigItem
//params cFileName the name of the file
func NewManager(cFileName string) (*Manager, error) {
	confs, err := lazyf.GetConfig(cFileName)
	if len(confs) == 0 {
		return &Manager{}, errors.New("Conf, completely empty")
	}
	if err != nil {
		return &Manager{}, errors.Wrap("Could not load conf", err)
	}
	cfig := confs[0]
	man = &Manager{
		filename: cFileName,
		rootLoc:  cfig.PString(cFilename, "root"),
		lastEdit: time.Now(),
	}
	for k, c := range confs[1] {
	}

}

//TryTemplate is the main useful method takes
//w: io writer
//host: the request.URL.Host
//p:the template name
//data:The data to send to the template
func (man *Manager) TryTemplate(w io.Writer, host string, p string, data interface{}) error {
	t, err := t.getTemplates(host)

	if err != nil {
		return errors.New("Could not access templates for :" + host)
	}

	b := new(bytes.Buffer)
	err = t.ExecuteTemplate(b, p, data)
	if err != nil {
		return err
	}
	io.Copy(w, b)

}

func (man *Manager) GetConfig(host string) (ConfigItem, error) {

	for _, v := range man.config {
		if v.CanHost(host) {
			return v, nil
		}
	}
	return ConfigItem{}, errors.New("No config assigned to that name")

}

func (man *Manager) GetFilePath(host, fname string) (string, error) {
	if host == "" {
		p := path.Join(man.root, rpath), nil
		if !strings.hasPrefix(p, man.root) {
			return "", errors.New("No upward pathing")
		}
		return p, nil
	}

	c, err := man.GetConfig(host)
	if err != nil {
		return "", errors.Wrap(err, "Could not get file path")
	}

	rpath = c.Folder
	if len(rpath) == 0 {
		return "", errors.New("No Folder location for host:" + host)
	}
	if c.Folder[0] != "/" {
		rpath = path.Join(man.root, rpath)
	}

	res := path.Join(rpath, fname)
	if !strings.HasPrefix(res, rpath) {
		return "", errors.New("No Upwards path building")
	}
	return res, nil

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
		ts, err := GetModified(man.filename)
		if err == nil {

			if ts.After(lastCheck) {
				Log("Config File Changed")
				newcon, err := loadConfig(man.filename)
				if err == nil {
					oldmap := man.tmap
					man.Lock()
					man.config = newcon
					man.tmap = newTMap(man.root, man.config)
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
			ts, err := GetModified(modpath)
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
	return c.plates, nil
}
