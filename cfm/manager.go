//configmanager is a holder for all the separate hosts and the folders they represent.
//A configuration reads a json file containing an array [] of ConfigItem s
//If the last 'Host' is 'default' this will be a catch all
package cfm

import (
	"path"
	"sync"

	"github.com/coderconvoy/lazyf"
	"github.com/pkg/errors"

	"time"
)

type Manager struct {
	filename   string
	rootLoc    string
	confs      lazyf.LZ
	sites      []ConfigItem
	sync.Mutex // Currently just for logger
}

func (m Manager) LogLoc() string {
	return path.Join(m.rootLoc, "logs")
}

//NewManager Creates a new Manager from json file based on ConfigItem
//params cFileName the name of the file
func NewManager(confs []lazyf.LZ, cFileName string) (*Manager, error) {

	if len(confs) == 0 {
		return &Manager{}, errors.New("Conf, completely empty")
	}
	if len(confs) == 1 {
		return &Manager{}, errors.New("Conf contains no Sites")
	}
	cfig := confs[0]
	man := &Manager{
		confs:    cfig,
		filename: cFileName,
		rootLoc:  cfig.PStringD(cFileName, "root"),
	}

	var err error = nil
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

//Note Locking Method for map safety. Use with some care
func (man *Manager) GetConfig(host string) (ConfigItem, error) {

	for _, v := range man.sites {
		if v.CanHost(host) {
			return v, nil
		}
	}
	return nil, errors.New("No config assigned to that name")

}

/**
func (man *Manager) GetFilePath(host, fname string) (string, error) {
	//Not looking for host
	if host == "" {
		return SafeJoin(man.rootLoc, fname)
	}

	c, err := man.GetConfig(host)
	if err != nil {
		return "", errors.Wrap(err, "Could not get file path")
	}
	return c.GetFilePath(fname)

}
*/

func manageTemplates(man *Manager) {

	for {

		//check folders for update only update the changed
		for _, v := range man.sites {
			v.Update()
		}
		//for each file look at modified file if changed update.
		time.Sleep(time.Minute / 3)
	}

}

/*
func (man *Manager) getTemplates(host string) (*tempower.PowerTemplate, error) {
	c, err := man.GetConfig(host)
	if err != nil {
		return nil, errors.Wrap(err, "No config available for host: "+host)
	}
	return c.Plates(), nil
}
*/

func (man *Manager) Confs() lazyf.LZ {
	return man.confs
}

func (man *Manager) Domains() []string {
	res := []string{}
	for _, v := range man.sites {
		res = append(res, v.Domains()...)
	}

	www := []string{}
	for _, v := range res {
		www = append(www, "www."+v)
	}

	return append(res, www...)
}
