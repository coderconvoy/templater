//configmanager is a holder for all the separate hosts and the folders they represent.
//A configuration reads a json file containing an array [] of ConfigItem s
//If the last 'Host' is 'default' this will be a catch all
package cfm

import (
	"bytes"
	"path"

	"github.com/coderconvoy/lazyf"
	"github.com/coderconvoy/templater/tempower"
	"github.com/pkg/errors"

	"io"
	"sync"
	"time"
)

type Manager struct {
	filename string
	rootLoc  string
	sites    []ConfigItem
	sync.Mutex
}

//NewManager Creates a new Manager from json file based on ConfigItem
//params cFileName the name of the file
func NewManager(cFileName string) (*Manager, error) {
	confs, err := lazyf.GetConfig(cFileName)
	if err != nil {
		return &Manager{}, errors.Wrap(err, "Could not load conf")
	}
	if len(confs) == 0 {
		return &Manager{}, errors.New("Conf, completely empty")
	}
	cfig := confs[0]
	man := &Manager{
		filename: cFileName,
		rootLoc:  cfig.PStringD(cFileName, "root"),
	}

	err = nil
	for _, c := range confs[1:] {
		nc, e := NewConfigItem(c, man.rootLoc)
		if e != nil {
			err = e
			continue
		}
		man.sites = append(man.sites, nc)
	}

	go manageTemplates(man)
	return man, err
}

//TryTemplate is the main useful method takes
//w: io writer
//host: the request.URL.Host
//p:the template name
//data:The data to send to the template
func (man *Manager) TryTemplate(w io.Writer, host string, p string, data interface{}) error {
	t, err := man.getTemplates(host)

	if err != nil {
		return errors.New("Could not access templates for :" + host)
	}

	b := new(bytes.Buffer)
	err = t.ExecuteTemplate(b, p, data)
	if err != nil {
		return err
	}
	io.Copy(w, b)
	return nil
}

func (man *Manager) GetConfig(host string) (ConfigItem, error) {

	man.Lock()
	defer man.Unlock()
	for _, v := range man.sites {
		if v.CanHost(host) {
			return v, nil
		}
	}
	return ConfigItem{}, errors.New("No config assigned to that name")

}

func (man *Manager) GetFilePath(host, fname string) (string, error) {
	//Not looking for host
	if host == "" {
		return SafeJoin(man.rootLoc, fname)
	}

	c, err := man.GetConfig(host)
	if err != nil {
		return "", errors.Wrap(err, "Could not get file path")
	}

	rpath := c.Folder
	if len(rpath) == 0 {
		return "", errors.New("No Folder location for host:" + host)
	}
	if c.Folder[0] != '/' {
		rpath = path.Join(man.rootLoc, rpath)
	}

	res, err := SafeJoin(rpath, fname)
	if err != nil {
		return "", err
	}
	return res, nil

}

func manageTemplates(man *Manager) {

	lastCheck := time.Now()
	var thisCheck time.Time

	for {
		thisCheck = time.Now()

		//check folders for update only update the changed
		for k, v := range man.sites {
			modpath := path.Join(v.Folder, v.Modifier)
			ts, err := GetModified(modpath)
			if err == nil {
				if ts.After(lastCheck) {
					fol := v.Folder
					newPlates, err := tempower.NewPowerTemplate(path.Join(fol, "templates/*"), fol)
					if err != nil {
						//TODO Log it
						continue
					}
					man.Lock()
					man.sites[k].plates = newPlates
					man.Unlock()
				}

			} else {
			}

		}
		//for each file look at modified file if changed update.
		lastCheck = thisCheck
		time.Sleep(time.Minute / 3)
	}

}

func (man *Manager) getTemplates(host string) (*tempower.PowerTemplate, error) {
	c, err := man.GetConfig(host)
	if err != nil {
		return nil, errors.Wrap(err, "No config available for host: "+host)
	}
	return c.plates, nil
}
