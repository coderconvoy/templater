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

type configItem struct {
	Host     string
	Folder   string
	Modifier string
}

type Manager struct {
	filename string
	tmap     map[string]*tempower.PowerTemplate
	config   []configItem
	killflag bool
	sync.Mutex
}

func newTMap(conf []configItem) map[string]*tempower.PowerTemplate {
	var err error
	res := make(map[string]*tempower.PowerTemplate)

	for _, v := range conf {
		t, ok := res[v.Folder]
		if !ok {
			t, err = tempower.NewPowerTemplate(path.Join(v.Folder, "templates/*.*"), v.Folder)
			if err == nil {
				res[v.Folder] = t
			}
		}
	}
	return res

}

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
		if err != nil {
			if fi.ModTime().After(lastCheck) {
				fmt.Println("Config File Changed")
				newcon, err := loadConfig(man.filename)
				if err != nil {
					oldmap := man.tmap
					man.Lock()
					man.config = newcon
					man.tmap = newTMap(man.config)
					man.Unlock()

					for _, v := range oldmap {
						v.Kill()
					}

				} else {
					//ignore the change
					fmt.Println("Load Config Error:", err)
				}
			}
		}

		//check folders for update only update the changed
		for _, _ := range man.config {

			//todo fix lots to loop on map of struct

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

func (man *Manager) getTemplate(rootF string) *tempower.PowerTemplate {
	man.Lock()
	defer man.Unlock()
	t, ok := man.tmap[rootF]
	if !ok {
		t, ok = man.tmap["default"]
	}
	return t

}

func (man *Manager) tryTemplate(w io.Writer, rootF string, p string, data interface{}) error {

	b := new(bytes.Buffer)
	for i := 0; i < 10; i++ {
		t := man.getTemplate(rootF)
		if t == nil {
			return fmt.Errorf("No Templates exist with this root:%f\n", rootF)
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

func (man *Manager) Kill() {
	man.Lock()
	defer man.Unlock()

	man.killflag = true
	//TODO loop through and kill all templates
	for _, v := range man.tmap {
		v.Kill()
	}
}
