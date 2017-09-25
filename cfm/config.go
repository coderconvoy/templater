package cfm

import (
	"errors"
	"path"
	"strings"
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
	plates   *tempower.PowerTemplate
	last     time.Time
}

func NewConfigItem(lz lazyf.LZ, root string) (ConfigItem, error) {
	fol, err := lz.PString("folder", "Folder")
	if err != nil || len(fol) == 0 {
		return ConfigItem{}, errors.New("No folder for item")
	}
	if fol[0] != '/' {
		fol = path.Join(root, fol)
	}

	item := ConfigItem{
		Hosts:    lz.PStringAr("host", "Host"),
		Folder:   fol,
		Modifier: lz.PStringD("Mod", "mod", "Modifier", "modifier"),
		last:     time.Now(),
	}
	plates, err := tempower.NewPowerTemplate(path.Join(fol, "templates/*"), fol)
	if err != nil {
		return ConfigItem{}, err
	}
	item.plates = plates

	return item, nil
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
