//configmanager is a holder for all the separate hosts and the folders they represent.
//A configuration reads a json file containing an array [] of ConfigItem s
package configmanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/coderconvoy/templater/blob"
	"github.com/coderconvoy/templater/tempower"
	"path"

	"io"
	"io/ioutil"
	"os"
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
	tmap     map[string]temroot
	config   []configItem
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
		err := t.ExecuteTemplate(b, p, data)
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
	t, err := tempower.NewPowerTemplate(path.Join(fol, "templates/*.*"), fol)
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

func newTMap(conf []configItem) map[string]temroot {
	res := make(map[string]temroot)

	for _, v := range conf {
		_, ok := res[v.Folder]
		if !ok {
			t, err := newTemroot(v.Folder, v.Modifier)
			if err == nil {
				res[v.Folder] = t
			} else {
				fmt.Printf("Could not load templates :%s,%s", v.Folder, err)
			}
		}
	}
	return res
}

func loadConfig(fName string) ([]configItem, error) {
	var configs []configItem

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

	var lastCheck time.Time
	var thisCheck time.Time

	for {
		thisCheck = time.Now()

		//if config has been updated then reset everything
		fi, err := os.Stat(man.filename)
		if err == nil {
			if fi.ModTime().After(lastCheck) {
				fmt.Println("Config File Changed")
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
					fmt.Println("Load Config Error:", err)
				}
			}
		}

		//check folders for update only update the changed
		for k, v := range man.tmap {
			modpath := path.Join(v.root, v.modifier)
			fi, err := os.Stat(modpath)
			if err == nil {
				if fi.ModTime().After(lastCheck) {
					t, err2 := newTemroot(v.root, v.modifier)
					if err2 == nil {
						man.Lock()
						man.tmap[k] = t
						v.templates.Kill()
						man.Unlock()
					} else {
						fmt.Printf("ERROR , Could not parse templates Using old ones: %s,%s\n", modpath, err2)
					}

				}

			} else {
				fmt.Printf("ERROR, Mod file missing:%s,%s\n ", modpath, err)
			}

		}

		//Allow kill
		if man.killflag {
			return
		}
		//for each file look at modified file if changed update.
		lastCheck = thisCheck
		time.Sleep(time.Minute)
	}

}

func (man *Manager) getTemplates(host string) (*tempower.PowerTemplate, error) {
	man.Lock()
	defer man.Unlock()
	for i, v := range man.config {
		if v.Host == host {
			t, ok := man.tmap[v.Folder]
			if ok {
				return t.templates, nil
			}
		}
	}
	t, ok := man.tmap["default"]
	if !ok {
		return nil, fmt.Errorf("No Templates available for host : %s\n", host)
	}
	return t.templates, nil

}
