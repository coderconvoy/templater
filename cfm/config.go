package cfm

import (
	"errors"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/coderconvoy/lazyf"
	"github.com/coderconvoy/templater/tempower"
)

type ConfigItem struct {
	//The host which will redirect to the folder
	Hosts []string
	//The Folder containing this hosting, -multiple hosts may point to the same folder
	Folder string
	//The Filename inside the Folder of the file we watch for changes
	Modifier string
	lastMod  time.Time
	plates   *tempower.PowerTemplate
	sync.Mutex
}

func NewConfigItem(lz lazyf.LZ, root string) (*ConfigItem, error) {
	fol, err := lz.PString("folder", "Folder")
	if err != nil || len(fol) == 0 {
		return nil, errors.New("No folder for item")
	}
	if fol[0] != '/' {
		fol = path.Join(root, fol)
	}

	item := &ConfigItem{
		Hosts:    lz.PStringAr("host", "Host"),
		Folder:   fol,
		Modifier: lz.PStringD("modify", "Mod", "mod", "Modifier", "modifier"),
	}
	item.lastMod = item.Mod()

	plates, err := tempower.NewPowerTemplate(path.Join(fol, "templates/*"), fol)
	if err != nil {
		return nil, err
	}
	item.plates = plates

	return item, nil
}

func (c ConfigItem) Plates() *tempower.PowerTemplate {
	c.Lock()
	res := c.plates
	c.Unlock()
	return res
}

func (c ConfigItem) CanHost(s string) bool {
	s = strings.TrimPrefix(s, "www.")
	for _, v := range c.Hosts {
		if s == v {
			return true
		}
		if v == "default" {
			return true
		}
	}
	return false
}

func (c *ConfigItem) Log(s string) {
	if len(c.Hosts) == 0 {
		Log(s)
	}
	LogTo(c.Hosts[0], s)
}

func (c *ConfigItem) Update() {
	nts := c.Mod()
	if nts.Equal(c.lastMod) {
		return
	}

	newPlates, err := tempower.NewPowerTemplate(path.Join(c.Folder, "templates/*"), c.Folder)
	if err != nil {
		c.Log("Could not Update templates:" + err.Error())
		return
	}

	c.Lock()
	c.plates = newPlates
	c.Unlock()
	c.lastMod = nts
}

func (c ConfigItem) Mod() time.Time {
	mpath := path.Join(c.Folder, c.Modifier)
	t, err := GetModified(mpath)
	if err != nil {
		return time.Time{}
	}
	return t
}
